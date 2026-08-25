// Package config holds the configuration for a single conversion run.
package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// Config represents the configuration for a single conversion run.
type Config struct {
	InputFile          string      `mapstructure:"input-file" validate:"required,file"`
	OutputFile         string      `mapstructure:"output-file" validate:"required,filepath"`
	ImagesDirs         Directories `mapstructure:",squash"`
	ExtractedImagesDir string      `mapstructure:"extracted-images-dir" validate:"omitempty,dir"`
	TemplateDirs       []string    `mapstructure:"template-dirs" validate:"omitempty"`
	Level              string      `mapstructure:"log-level" validate:"omitempty,oneof=debug info warn error"`
	Concurrency        int         `mapstructure:"concurrency" validate:"omitempty,min=1,max=128"`
}

// Validate validates the Config using the validator library.
// It returns an error if the Config is invalid.
func (c Config) Validate() error {
	validate := validator.New()

	// Tell validator to read mapstructure tags instead of Go struct field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("mapstructure"), ",")
		if name == "-" {
			return ""
		}
		return name
	})

	// Set up the English translator
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(validate, trans)

	err := validate.Struct(c)
	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		var errMsgs []string
		for _, e := range errs {
			errMsgs = append(errMsgs, e.Translate(trans))
		}
		return fmt.Errorf("configuration validation failed:\n  - %s", strings.Join(errMsgs, "\n  - "))
	}

	return nil
}
