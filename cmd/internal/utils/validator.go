package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func GetJSONFieldName(err validator.FieldError) string {
	return err.Field()
}

func ValidateBodyJSON[T any](c *gin.Context, body *T) error {
	if err := c.ShouldBindJSON(body); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {

			formattedErrors := make(map[string]string)
			for _, f := range errs {
				formattedErrors[f.Field()] = f.Tag()
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": formattedErrors})
			return err
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return err
	}

	return nil
}
