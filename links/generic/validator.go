package generic

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validator = validator.New()

var (
	handleRegex           = regexp.MustCompile(`^[a-z0-9]{3,24}$`)
	blacklistedCSSStrings = [...]string{
		// php strings
		"<?php", "?>", ".php",
		// javascript, remember Samy from myspace :)
		"javascript:eval", ":eval", "expr=", ".expr", "script", "javascript", "javascript:",
		// IE javascript
		"expression",
		// capture input value, we won't have input but just in case
		"[value",
		// https://stackoverflow.com/questions/476276/using-javascript-in-css
		".htc", ".xml", "behavior", "binding",
	}
)

func init() {
	if err := Validator.RegisterValidation("handle", func(field validator.FieldLevel) bool {
		i := field.Field().Interface()
		s, ok := i.(string)
		if !ok {
			return false
		}

		return handleRegex.MatchString(s)
	}); err != nil {
		panic(err)
	}

	if err := Validator.RegisterValidation("css", func(field validator.FieldLevel) bool {
		i := field.Field().Interface()
		s, ok := i.(string)
		if !ok {
			return false
		}

		for _, bs := range blacklistedCSSStrings {
			if strings.Contains(s, bs) {
				return false
			}
		}

		return true
	}); err != nil {
		panic(err)
	}
}
