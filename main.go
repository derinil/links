package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/derinil/links/links/account"
	"github.com/derinil/links/links/account/auth"
	"github.com/derinil/links/links/account/auth/handlers"
	"github.com/derinil/links/links/account/session"
	"github.com/derinil/links/links/cache"
	"github.com/derinil/links/links/crypto/csrf"
	"github.com/derinil/links/links/database"
	"github.com/derinil/links/links/database/migrator"
	"github.com/derinil/links/links/generic"
	"github.com/derinil/links/links/views"
	"github.com/derinil/links/links/web"
	"github.com/derinil/links/links/web/responder"
	"github.com/derinil/links/migrations"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Environment string `required:"true"`
	Database    struct {
		MaxConns int    `default:"100"`
		DSN      string `required:"true"`
	}
	Redis struct {
		Address  string `required:"true"`
		Password string
	}
	Server struct {
		RequestTimeout    time.Duration `default:"30s"`
		ReadHeaderTimeout time.Duration `default:"5s"`
	}
	Secrets struct {
		CSRFKey []byte `split_words:"true" required:"true"`
	}
}

func main() {
	_ = godotenv.Load()

	var cfg config
	if err := envconfig.Process("links", &cfg); err != nil {
		log.Fatalln("failed to load config", err)
	}

	if err := runMigrations(&cfg); err != nil {
		log.Fatalln("failed to run migrations", err)
	}

	log.Println("migrations finished running")

	if err := runServer(&cfg); err != nil {
		log.Fatalln("failed to run server", err)
	}
}

func runServer(cfg *config) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	db, err := database.ConnectWithContext(ctx, cfg.Database.DSN, cfg.Database.MaxConns)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalln("failed to close database", err)
		}
	}()

	rds, err := cache.NewRedis(ctx, cfg.Redis.Address, cfg.Redis.Password)
	if err != nil {
		return fmt.Errorf("failed to open redis: %w", err)
	}

	defer func() {
		if err := rds.Close(); err != nil {
			log.Fatalln("failed to close redis", err)
		}
	}()

	var (
		accountReader = database.NewAccountReader(db)
		accountWriter = database.NewAccountWriter(db)
	)

	var (
		sessionHandler = session.NewHandler(rds)
		csrfHandler    = csrf.NewHandler(cfg.Secrets.CSRFKey)
		viewsHandler   = views.NewHandler(
			views.IndexPageRenderer(),
			views.LoginPageRenderer(),
			views.LinksPageRenderer(),
			views.AccountPageRenderer(),
			views.RegisterPageRenderer(),
		)
		accountHandler = account.NewHandler(accountReader, accountWriter)
		authHandler    = auth.NewHandler(
			handlers.LogoutHandler(sessionHandler),
			handlers.LoginHandler(accountHandler, sessionHandler),
			handlers.RegistrationHandler(accountHandler, sessionHandler),
		)
	)

	var (
		responderHandler = responder.NewHandler()
		webHandler       = web.NewHandler(
			authHandler,
			csrfHandler,
			viewsHandler,
			accountHandler,
			sessionHandler,
			responderHandler,
		)

		router = chi.NewMux()
		server = &http.Server{
			Addr:              ":8080",
			ReadTimeout:       cfg.Server.RequestTimeout,
			WriteTimeout:      cfg.Server.RequestTimeout,
			IdleTimeout:       cfg.Server.RequestTimeout,
			ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
			Handler:           router,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
		}
	)

	router.Use(generic.RequestBeginTime)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	if cfg.Environment == "local" {
		router.Use(middleware.NoCache)
	}
	router.Use(middleware.Timeout(cfg.Server.RequestTimeout))

	if cfg.Environment == "local" {
		router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./links/views/static"))))
	} else {
		router.Handle("/static/*", http.FileServer(http.FS(views.StaticFiles)))
	}

	router.Mount("/", webHandler.Router())

	go server.ListenAndServe()

	log.Println("running on port :8080")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-quit
	cancel()

	return nil
}

func runMigrations(cfg *config) error {
	ctx := context.Background()

	db, err := database.ConnectWithContext(ctx, cfg.Database.DSN, cfg.Database.MaxConns)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	defer func() {
		if err = db.Close(); err != nil {
			log.Fatalln("failed to close database", err)
		}
	}()

	m := migrator.New(migrations.MigrationsFS, ".", db)

	if err := m.Up(); err != nil {
		return fmt.Errorf("failed to up migrations: %w", err)
	}

	return nil
}
