package validation

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

func (ve ValidationErrors) Error() string {
	var log []string
	for _, v := range ve {
		log = append(log, v.Error())
	}
	return strings.Join(log, ";")
}
