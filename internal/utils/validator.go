package utils

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

func InitValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("password", validatePassword)
	return validate
}

func InitUniversalTranslator(validate *validator.Validate) *ut.UniversalTranslator {
	en := en.New()

	uni := ut.New(en, en)

	return uni
}

func InitTranslator(validate *validator.Validate, uni *ut.UniversalTranslator) ut.Translator {
	trans, _ := uni.GetTranslator("en")

	en_translations.RegisterDefaultTranslations(validate, trans)

	validate.RegisterTranslation("password", trans, func(ut ut.Translator) error {
		return ut.Add("password", "password should contain at least 1 uppercase, 1 lowercase, 1 number and 1 punctuation/symbol", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("password", fe.Field())
		return t
	})

	return trans
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var (
		hasUpper  = false
		hasLower  = false
		hasNumber = false
		hasSymbol = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSymbol
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	parts := strings.Split(jsonTag, ",")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return field.Name
}

func ValidateRequest[T any](validate *validator.Validate, trans ut.Translator, req T) []ValidationError {
	err := validate.Struct(req)
	if err != nil {
		var errors []ValidationError

		for _, err := range err.(validator.ValidationErrors) {
			field, _ := reflect.TypeOf(req).FieldByName(err.Field())
			jsonFieldName := getJSONFieldName(field)
			var message string
			switch err.Tag() {
			case "required":
				message = fmt.Sprintf("%s is required", jsonFieldName)
			case "email":
				message = err.Translate(trans)
			case "min":
				message = fmt.Sprintf("%s must be at least %s characters long", jsonFieldName, err.Param())
			case "max":
				message = fmt.Sprintf("%s must be at most %s characters long", jsonFieldName, err.Param())
			case "len":
				message = fmt.Sprintf("%s must be %s characters long", jsonFieldName, err.Param())
			case "alphanum":
				message = fmt.Sprintf("%s must be alphanumeric characters only", jsonFieldName)
			case "password":
				message = err.Translate(trans)
			default:
				message = fmt.Sprintf("%s is invalid", jsonFieldName)
			}

			errors = append(errors, ValidationError{
				Field:   jsonFieldName,
				Message: message,
			})
		}
		return errors
	}
	return nil
}
