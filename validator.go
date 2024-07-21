package main

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

func InitValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

func InitUniversalTranslator(validate *validator.Validate) *ut.UniversalTranslator {
	en := en.New()

	uni := ut.New(en, en)

	return uni
}

func InitTranslator(validate *validator.Validate, uni *ut.UniversalTranslator) ut.Translator {
	trans, _ := uni.GetTranslator("en")

	en_translations.RegisterDefaultTranslations(validate, trans)

	return trans
}
