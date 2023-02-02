package web

import (
	"net/http"

	"github.com/derinil/links/links/account"
	"github.com/derinil/links/links/account/auth"
	"github.com/derinil/links/links/account/auth/handlers"
	"github.com/derinil/links/links/account/session"
	"github.com/derinil/links/links/crypto/csrf"
	"github.com/derinil/links/links/views"
	"github.com/derinil/links/links/web/responder"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	authHandler      auth.Handler
	csrfHandler      csrf.Handler
	viewsHandler     views.Handler
	accountHandler   account.Handler
	sessionHandler   session.Handler
	responderHandler responder.Handler
}

func NewHandler(
	authHandler auth.Handler,
	csrfHandler csrf.Handler,
	viewsHandler views.Handler,
	accountHandler account.Handler,
	sessionHandler session.Handler,
	responderHandler responder.Handler,
) *Handler {
	return &Handler{
		authHandler:      authHandler,
		csrfHandler:      csrfHandler,
		viewsHandler:     viewsHandler,
		accountHandler:   accountHandler,
		sessionHandler:   sessionHandler,
		responderHandler: responderHandler,
	}
}

func (s *Handler) Router() *chi.Mux {
	var (
		r = chi.NewMux()

		parseSession   = session.ParseSession(s.sessionHandler)
		forceSession   = session.ForceSession(s.responderHandler)
		forceNoSession = session.ForceNoSession(s.responderHandler)
		injectCSRF     = csrf.InjectCSRF(s.csrfHandler)
		validateCSRF   = csrf.ValidateCSRF(s.csrfHandler, s.responderHandler)
	)

	r.Use(parseSession)
	r.Use(injectCSRF)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:     "/",
			ErrorMsg: "Not Found!",
		})
	})

	// Authenticated account pages
	r.With(forceSession).Group(func(r chi.Router) {
		r.Route("/account", func(r chi.Router) {
			// Account page
			r.Get("/", s.renderAccountPage)
			// Update account
			r.With(validateCSRF).Post("/", s.handleUpdateAccount)
		})

		// Log out
		r.With(validateCSRF).Post("/logout", s.handleLogout)
	})

	// Unauthenticated pages like /register and /login
	r.With(forceNoSession).Group(func(r chi.Router) {
		// GET forms
		r.Get("/register", s.genericRenderPage(views.Register))
		r.Get("/login", s.genericRenderPage(views.Login))

		// POST forms
		r.With(validateCSRF).Group(func(r chi.Router) {
			r.Post("/register", s.handleRegistration)
			r.Post("/login", s.handleLogin)
		})
	})

	// Index page
	r.Get("/", s.genericRenderPage(views.Index))

	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Links page for a user
	r.Get("/{handle}", s.renderLinksPage)

	return r
}

func (s *Handler) genericRenderPage(page views.Page) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.viewsHandler.Render(r.Context(), w, page, &views.RenderCmd{
			Error:   r.URL.Query().Get("error"),
			Message: r.URL.Query().Get("message"),
		})
	}
}

func (s *Handler) handleRegistration(w http.ResponseWriter, r *http.Request) {
	var (
		f   = r.Form
		ctx = r.Context()
		cmd = &handlers.RegistrationCmd{
			Name:     f.Get("name"),
			Handle:   f.Get("handle"),
			Password: f.Get("password"),
		}
	)

	a, err := s.authHandler.Handle(ctx, &auth.AuthCmd{
		Method: auth.Register,
		Cmd:    cmd,
	})
	if err != nil {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/register",
			Error: err,
		})
		return
	}

	http.SetCookie(w, session.Cookie(a.SessionToken))

	s.responderHandler.Respond(w, r, &responder.ResponseCmd{
		Path:    "/account",
		Message: "Successfully registered!",
	})
}

func (s *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var (
		f   = r.Form
		ctx = r.Context()
		cmd = &handlers.LoginCmd{
			Handle:   f.Get("handle"),
			Password: f.Get("password"),
		}
	)

	a, err := s.authHandler.Handle(ctx, &auth.AuthCmd{
		Method: auth.Login,
		Cmd:    cmd,
	})
	if err != nil {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/login",
			Error: err,
		})
		return
	}

	http.SetCookie(w, session.Cookie(a.SessionToken))

	s.responderHandler.Respond(w, r, &responder.ResponseCmd{
		Path:    "/account",
		Message: "Successfully logged in!",
	})
}

func (s *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var (
		ctx = r.Context()
		cmd = &handlers.LogoutCmd{}
	)

	st, ok := ctx.Value(session.SessionTokenKey).(string)
	if !ok {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/",
			Error: session.ErrNotAuthenticated,
		})
		return
	}

	cmd.SessionToken = st

	_, err := s.authHandler.Handle(ctx, &auth.AuthCmd{
		Method: auth.Logout,
		Cmd:    cmd,
	})
	if err != nil {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/",
			Error: err,
		})
		return
	}

	http.SetCookie(w, session.RemoveCookie())

	s.responderHandler.Respond(w, r, &responder.ResponseCmd{
		Path:    "/",
		Message: "Successfully logged out!",
	})
}

func (s *Handler) renderAccountPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	so, ok := ctx.Value(session.SessionObjectKey).(*session.Session)
	if !ok {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/login",
			Error: session.ErrNotAuthenticated,
		})
		return
	}

	a, err := s.accountHandler.Get(ctx, &account.GetCmd{ID: so.AccountID})
	if err != nil {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/",
			Error: err,
		})
		return
	}

	if len(a.Links) == 0 {
		a.Links = append(a.Links, account.Link{
			Title: "My Github Link!",
			Link:  "http://github.com",
		})
	}

	s.viewsHandler.Render(r.Context(), w, views.Account, &views.RenderCmd{
		Error:   r.URL.Query().Get("error"),
		Message: r.URL.Query().Get("message"),
		Cmd:     views.AccountPageCmd{Account: a},
	})
}

func (s *Handler) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	var (
		f   = r.Form
		ctx = r.Context()
		cmd = &account.UpdateCmd{
			Name:   f.Get("name"),
			Handle: f.Get("handle"),
			CSS:    f.Get("css"),
		}
	)

	if len(f["links_title[]"]) > 0 && len(f["links_title[]"]) == len(f["links_url[]"]) {
		for i := range f["links_title[]"] {
			cmd.Links = append(cmd.Links, account.LinkScaffold{
				Title: f["links_title[]"][i],
				Link:  f["links_url[]"][i],
			})
		}
	}

	so, ok := ctx.Value(session.SessionObjectKey).(*session.Session)
	if !ok {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/login",
			Error: session.ErrNotAuthenticated,
		})
		return
	}

	cmd.AccountID = so.AccountID

	_, err := s.accountHandler.Update(ctx, cmd)
	if err != nil {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/account",
			Error: err,
		})
		return
	}

	s.responderHandler.Respond(w, r, &responder.ResponseCmd{
		Path:    "/account",
		Message: "Successfully updated account information!",
	})
}

func (s *Handler) renderLinksPage(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    = r.Context()
		handle = chi.URLParam(r, "handle")
	)

	a, err := s.accountHandler.Get(ctx, &account.GetCmd{Handle: handle})
	if err != nil {
		s.responderHandler.Respond(w, r, &responder.ResponseCmd{
			Path:  "/",
			Error: err,
		})
		return
	}

	s.viewsHandler.Render(r.Context(), w, views.Links, &views.RenderCmd{
		Cmd: &views.LinksPageCmd{Account: a},
	})
}
