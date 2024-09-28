package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type ValidationService struct {
	validate *validator.Validate
	uni      *ut.UniversalTranslator
	trans    ut.Translator
}

type Validator interface {
	Validate(T any) []ValidationError
}

func NewValidationService() *ValidationService {
	validate := initValidator()
	uni := initUniversalTranslator()
	trans := initTranslator(validate, uni)

	return &ValidationService{
		validate: validate,
		uni:      uni,
		trans:    trans,
	}
}
func (s *ValidationService) Validate(data any) []ValidationError {
	err := s.validate.Struct(data)
	if err != nil {
		var errors []ValidationError

		for _, err := range err.(validator.ValidationErrors) {
			field, _ := reflect.TypeOf(data).FieldByName(err.Field())
			jsonFieldName := getJSONFieldName(field)
			var message string
			switch err.Tag() {
			case "required":
				message = fmt.Sprintf("%s is required", jsonFieldName)
			case "email":
				message = err.Translate(s.trans)
			case "min":
				message = fmt.Sprintf("%s must be at least %s characters long", jsonFieldName, err.Param())
			case "max":
				message = fmt.Sprintf("%s must be at most %s characters long", jsonFieldName, err.Param())
			case "len":
				message = fmt.Sprintf("%s must be %s characters long", jsonFieldName, err.Param())
			case "alphanum":
				message = fmt.Sprintf("%s must be alphanumeric characters only", jsonFieldName)
			case "password":
				message = err.Translate(s.trans)
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

func initValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("password", validatePassword)
	return validate
}

func initUniversalTranslator() *ut.UniversalTranslator {
	en := en.New()
	uni := ut.New(en, en)
	return uni
}

func initTranslator(validate *validator.Validate, uni *ut.UniversalTranslator) ut.Translator {
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

func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	parts := strings.Split(jsonTag, ",")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return field.Name
}
