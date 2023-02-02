package generic

import "errors"

type WebError struct {
	StatusCode int
	ErrKey     string
	ErrMsg     string
}

func NewWebError(statusCode int, key, msg string) *WebError {
	return &WebError{
		StatusCode: statusCode,
		ErrKey:     key,
		ErrMsg:     msg,
	}
}

func (e *WebError) Error() string {
	return e.ErrMsg
}

// Unwrap returns the unwrapped error and if that's nil,
// it returns the error back
func Unwrap(err error) error {
	e := err

	for e != nil {
		next := errors.Unwrap(e)
		if next == nil {
			return e
		}
		e = next
	}

	return e
}
