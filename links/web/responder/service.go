package responder

import (
	"log"
	"net/http"
	"net/url"

	"github.com/derinil/links/links/generic"
	"github.com/go-playground/validator/v10"
)

type (
	Handler interface {
		Respond(w http.ResponseWriter, r *http.Request, cmd *ResponseCmd)
	}

	HandlerImpl struct{}

	ResponseCmd struct {
		// Error and message will be passed in
		// the url query parameters when redirecting
		Error    error
		Message  string
		ErrorMsg string
		Path     string
	}
)

var _ Handler = (*HandlerImpl)(nil)

func NewHandler() *HandlerImpl {
	return &HandlerImpl{}
}

func (s *HandlerImpl) Respond(w http.ResponseWriter, r *http.Request, cmd *ResponseCmd) {
	errorMsg := cmd.ErrorMsg
	if err := generic.Unwrap(cmd.Error); err != nil {
		switch v := err.(type) {
		case *generic.WebError:
			errorMsg = v.ErrMsg
		case validator.FieldError:
			errorMsg = "Data is invalid"
		case validator.ValidationErrors:
			errorMsg = "Data is invalid"
		case *validator.InvalidValidationError:
			errorMsg = "Data is invalid"
		default:
			log.Println("unexpected error!", err)
			errorMsg = "Internal error, contact us!"
		}
	}

	v := url.Values{}
	if errorMsg != "" {
		v.Set("error", errorMsg)
	}
	if cmd.Message != "" {
		v.Set("message", cmd.Message)
	}

	path := cmd.Path
	if e := v.Encode(); e != "" {
		path += "?" + v.Encode()
	}

	http.Redirect(w, r, path, http.StatusFound)
}
