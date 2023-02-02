package views

import (
	"context"
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/derinil/links/links/account"
	"github.com/derinil/links/links/account/session"
	"github.com/derinil/links/links/crypto/csrf"
	"github.com/derinil/links/links/generic"
)

//go:embed *.html
var files embed.FS

//go:embed static/*
var StaticFiles embed.FS

type (
	// This handler will panic at any error as it does not rely
	// on user input and if anything goes wrong it is a crucial
	// error in the code and must be fixed immediately.
	Handler interface {
		Render(context.Context, http.ResponseWriter, Page, *RenderCmd)
	}

	HandlerImpl struct {
		pages map[Page]Renderer
	}

	Page string

	Renderer interface {
		Page() Page
		Render(http.ResponseWriter, *internalCmd)
	}

	RendererImpl struct {
		page   Page
		handle func(http.ResponseWriter, *internalCmd)
	}

	RenderCmd struct {
		Cmd     any
		Error   string
		Message string
	}

	internalCmd struct {
		// Cmd will be parsed/used by the specific template underlying cmd is made for
		Cmd any
		// These will be populated by default for each request and used by all templates
		Authenticated bool
		Handle        string
		ErrorMsg      string
		Message       string
		CSRFToken     string
		Took          time.Duration
	}

	AccountPageCmd struct {
		Account *account.Account
	}

	LinksPageCmd struct {
		Account *account.Account
	}
)

const (
	Index    Page = "index"
	Login    Page = "login"
	Links    Page = "links"
	Register Page = "register"
	Account  Page = "account"
)

var _ Handler = (*HandlerImpl)(nil)

func NewHandler(rs ...Renderer) *HandlerImpl {
	m := make(map[Page]Renderer, len(rs))
	for i := range rs {
		m[rs[i].Page()] = rs[i]
	}

	return &HandlerImpl{pages: m}
}

// It's ok to panic in this function as if an error occurs at this level, it is due to our
// own mishandling of the templates and cmds and this should result in a critical error.
func (s *HandlerImpl) Render(ctx context.Context, w http.ResponseWriter, p Page, cmd *RenderCmd) {
	var (
		r  = s.pages[p]
		rc = &internalCmd{}
	)

	rc.Took = time.Since(ctx.Value(generic.RequestBeginTimeKey).(time.Time))

	if v, ok := ctx.Value(session.SessionObjectKey).(*session.Session); ok {
		rc.Authenticated = true
		rc.Handle = v.Handle
	}

	if v, ok := ctx.Value(csrf.TokenKey).(string); ok {
		rc.CSRFToken = v
	}

	if cmd != nil {
		rc.Cmd = cmd.Cmd
		rc.ErrorMsg = cmd.Error
		rc.Message = cmd.Message
	}

	r.Render(w, rc)
}

func (s *RendererImpl) Page() Page {
	return s.page
}

func (s *RendererImpl) Render(w http.ResponseWriter, rc *internalCmd) {
	s.handle(w, rc)
}

// NOTE: If we end up with too much boilerplate code wrt RendererImpl
// we can make a generic renderer.

func IndexPageRenderer() *RendererImpl {
	tmpl := template.Must(template.ParseFS(files, "base.html", "index.html"))

	return &RendererImpl{
		page: Index,
		handle: func(w http.ResponseWriter, rc *internalCmd) {
			tmpl.Execute(w, rc)
		},
	}
}

func RegisterPageRenderer() *RendererImpl {
	tmpl := template.Must(template.ParseFS(files, "base.html", "register.html"))

	return &RendererImpl{
		page: Register,
		handle: func(w http.ResponseWriter, rc *internalCmd) {
			tmpl.Execute(w, rc)
		},
	}
}

func LoginPageRenderer() *RendererImpl {
	tmpl := template.Must(template.ParseFS(files, "base.html", "login.html"))

	return &RendererImpl{
		page: Login,
		handle: func(w http.ResponseWriter, rc *internalCmd) {
			tmpl.Execute(w, rc)
		},
	}
}

func LinksPageRenderer() *RendererImpl {
	tmpl := template.Must(template.ParseFS(files, "base.html", "links.html"))

	return &RendererImpl{
		page: Links,
		handle: func(w http.ResponseWriter, rc *internalCmd) {
			tmpl.Execute(w, rc)
		},
	}
}

func AccountPageRenderer() *RendererImpl {
	var (
		funcs = template.FuncMap{
			"add": func(x, y int) int {
				return x + y
			},
		}
		tmpl = template.Must(template.New("").Funcs(funcs).ParseFS(files, "base.html", "account.html"))
	)

	return &RendererImpl{
		page: Account,
		handle: func(w http.ResponseWriter, rc *internalCmd) {
			tmpl.ExecuteTemplate(w, "base.html", rc)
		},
	}
}
