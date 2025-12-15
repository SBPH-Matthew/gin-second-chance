package utils

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

func ValidationErrors(err error) map[string]string {
	errors := map[string]string{}

	for _, fieldErr := range err.(validator.ValidationErrors) {
		field := fieldErr.Field()

		switch fieldErr.Tag() {
		case "required":
			errors[field] = field + " is required"
		case "email":
			errors[field] = "Invalid email format"
		case "min":
			errors[field] = field + " must be at least " + fieldErr.Param() + " characters"
		default:
			errors[field] = "Invalid field"
		}
	}

	return errors
}
