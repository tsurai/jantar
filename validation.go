package amber

import (
	"time"
	"reflect"
)

type Validation struct {
	results []*ValidationResult
}

type ValidationResult struct {
	valid 	bool
	message	string
}

func (v *Validation) addValidationResult(valid bool, message string) *ValidationResult {
		result := &ValidationResult{valid, message}
		v.results = append(v.results, result)

		return result
}

func (vr *ValidationResult) Message(msg string) {
	if vr == nil {
		logger.Println("Failed to set validation Message. Result is nil")
	}

	vr.message = msg
}

func (v *Validation) Required(obj interface{}) *ValidationResult {
	defaultMessage := "Required"

	if obj == nil {
		return v.addValidationResult(false, defaultMessage)
	}
	
	if val, ok := obj.(int); ok {
		return v.addValidationResult(val != 0, defaultMessage)
	}

	if val, ok := obj.(string); ok {
		return v.addValidationResult(len(val) > 0, defaultMessage)
	}

	if val, ok := obj.(time.Time); ok {
		return v.addValidationResult(val.IsZero(), defaultMessage)
	}

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Slice {
		return v.addValidationResult(val.Len() > 0, defaultMessage)
	}

	return nil
}