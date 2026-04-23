package utils

import (
    "reflect"
    "strings"

    "github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func init() {
    Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })
}