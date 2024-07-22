package main

import (
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
