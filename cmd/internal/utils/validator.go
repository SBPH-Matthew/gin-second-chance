package utils

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"

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

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid form data", "errors": formattedErrors})
			return err
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON format"})
		return err
	}

	return nil
}

func ValidateBodyFormData(c *gin.Context, target interface{}) error {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Failed to parse multipart form"})
		return err
	}

	// Populate struct from form data
	if err := populateStructFromForm(c, target); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Failed to process form data"})
		return err
	}

	// Validate the populated struct
	if err := Validate.Struct(target); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			formattedErrors := make(map[string]string)
			for _, f := range errs {
				fieldName := strings.ToLower(f.Field())
				formattedErrors[fieldName] = getValidationMessage(f)
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Validation failed", "errors": formattedErrors})
			return err
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid form data"})
		return err
	}

	return nil
}

func populateStructFromForm(c *gin.Context, target interface{}) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil // Not a pointer to struct
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// Get form tag, fallback to lowercase field name if no tag
		formTag := fieldType.Tag.Get("form")
		var fieldName string
		if formTag != "" && formTag != "-" {
			// Use form tag, but split on comma to get the first part (ignore options like "omitempty")
			fieldName = strings.Split(formTag, ",")[0]
		} else {
			fieldName = strings.ToLower(fieldType.Name)
		}

		// Get form value
		formValue := c.PostForm(fieldName)
		if formValue == "" {
			continue
		}

		// Set field value based on type
		switch field.Kind() {
		case reflect.String:
			field.SetString(formValue)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if intVal, err := strconv.ParseInt(formValue, 10, 64); err == nil {
				field.SetInt(intVal)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if uintVal, err := strconv.ParseUint(formValue, 10, 64); err == nil {
				field.SetUint(uintVal)
			}
		case reflect.Float32, reflect.Float64:
			if floatVal, err := strconv.ParseFloat(formValue, 64); err == nil {
				field.SetFloat(floatVal)
			}
		case reflect.Bool:
			if boolVal, err := strconv.ParseBool(formValue); err == nil {
				field.SetBool(boolVal)
			}
		}
	}

	return nil
}

func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Minimum value/length is " + err.Param()
	case "max":
		return "Maximum value/length is " + err.Param()
	case "email":
		return "Invalid email format"
	default:
		return "Validation failed for " + err.Tag()
	}
}
