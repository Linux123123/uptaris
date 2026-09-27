package request

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/uptaris/uptaris/backend/internal/models"
)

type ValidationError struct {
	Issues []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Issues, "; ")
}

func (e *ValidationError) ValidationMessage() string { return e.Error() }

var checker = newValidator()
var translator = newTranslator(checker)

func newValidator() *validator.Validate {
	check := validator.New(validator.WithRequiredStructEnabled())
	check.SetTagName("binding")
	check.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		if name != "" {
			return name
		}

		return field.Name
	})
	check.RegisterStructValidation(func(level validator.StructLevel) {
		incident := level.Current().Interface().(models.Incident)
		if incident.ResolvedAt != nil && incident.ResolvedAt.Before(incident.StartedAt) {
			level.ReportError(incident.ResolvedAt, "resolvedAt", "ResolvedAt", "gtefield", "startedAt")
		}
	}, models.Incident{})

	return check
}

func newTranslator(check *validator.Validate) ut.Translator {
	english := en.New()
	universal := ut.New(english, english)
	translate, _ := universal.GetTranslator("en")
	if err := translations.RegisterDefaultTranslations(check, translate); err != nil {
		panic(err)
	}

	return translate
}

func Validate(value any) error {
	err := checker.Struct(value)
	if err == nil {
		return nil
	}

	fields, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	issues := make([]string, 0, len(fields))
	for _, field := range fields {
		issues = append(issues, field.Translate(translator))
	}

	return &ValidationError{Issues: issues}
}
