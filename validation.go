package amber

import (
	"fmt"
	"math"
	"time"
	"regexp"
	"reflect"
)

type Validation struct {
	hasErrors	bool
	errors 		map[string][]string
}

type ValidationResult struct {
	validation 	*Validation
	valid 			bool
	name				string
	index				int
}

func (v *Validation) HasErrors() bool {
	return v.hasErrors
}

func (v *Validation) addValidationResult(name string, valid bool, message string) *ValidationResult {
	result := &ValidationResult{v, valid, name, -1}

		if !valid {
			v.hasErrors = true
			v.errors[name] = append(v.errors[name], message)
			result.index = len(v.errors[name]) - 1
		}
		
		return result
}

func (vr *ValidationResult) Message(msg string) *ValidationResult {
	if vr == nil {
		logger.Println("[Warning] Failed to set validation message, result is nil")
	}

	if vr.index != -1 {
		vr.validation.errors[vr.name][vr.index] = msg
	}

	return vr
}

func (vr *ValidationResult) IsValid() bool {
	return vr.valid
}

func (v *Validation) Required(name string, obj interface{}) *ValidationResult {
	defaultMessage := "Required"

	if obj == nil {
		return v.addValidationResult(name, false, defaultMessage)
	}
	
	if value, ok := obj.(int); ok {
		return v.addValidationResult(name, value != 0, defaultMessage)
	}

	if value, ok := obj.(string); ok {
		return v.addValidationResult(name, len(value) > 0, defaultMessage)
	}

	if value, ok := obj.(time.Time); ok {
		return v.addValidationResult(name, value.IsZero(), defaultMessage)
	}

	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Slice {
		return v.addValidationResult(name, value.Len() > 0, defaultMessage)
	}

	return nil
}

func (v *Validation) Min(name string, obj interface{}, min int) *ValidationResult {
	defaultMessage := fmt.Sprintf("Must be larger than %d", min)

	if obj == nil {
		return v.addValidationResult(name, false, defaultMessage)
	}

	if value, ok := obj.(int); ok {
		return v.addValidationResult(name, value >= min, defaultMessage)
	}

	if value, ok := obj.(string); ok {
		return v.addValidationResult(name, len(value) >= min, defaultMessage)
	}

	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Slice {
		return v.addValidationResult(name, value.Len() >= min, defaultMessage)
	}

	return nil
}

func (v *Validation) Max(name string, obj interface{}, max int) *ValidationResult {
	defaultMessage := fmt.Sprintf("Must be smaller than %d", max)

	if obj == nil {
		return v.addValidationResult(name, false, defaultMessage)
	}

	if value, ok := obj.(int); ok {
		return v.addValidationResult(name, value <= max, defaultMessage)
	}

	if value, ok := obj.(string); ok {
		return v.addValidationResult(name, len(value) <= max, defaultMessage)
	}

	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Slice {
		return v.addValidationResult(name, value.Len() <= max, defaultMessage)
	}

	return nil
}

func (v *Validation) MinMax(name string, obj interface{}, min int, max int) *ValidationResult {
	defaultMessage := fmt.Sprintf("Must be larger %d and smaller %d", min, max)

	if obj == nil {
		return v.addValidationResult(name, false, defaultMessage)
	}

	if value, ok := obj.(int); ok {
		return v.addValidationResult(name, value >= min && value <= max, defaultMessage)
	}

	if value, ok := obj.(string); ok {
		return v.addValidationResult(name, len(value) >= min && len(value) <= max, defaultMessage)
	}

	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Slice {
		return v.addValidationResult(name, value.Len() >= min && value.Len() <= max, defaultMessage)
	}

	return nil
}

func (v *Validation) Length(name string, obj interface{}, length int) *ValidationResult {
	defaultMessage := fmt.Sprintf("Must be %d symbols long", length)

	if obj == nil {
		return v.addValidationResult(name, false, defaultMessage)
	}

	if value, ok := obj.(int); ok {
		return v.addValidationResult(name, int(math.Ceil(math.Log10(float64(value)))) == length, defaultMessage)
	}

	if value, ok := obj.(string); ok {
		return v.addValidationResult(name, len(value) == length, defaultMessage)
	}

	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Slice {
		return v.addValidationResult(name, value.Len() == length, defaultMessage)
	}

	return nil
}

func (v *Validation) Equals(name string, obj interface{}, obj2 interface{}) *ValidationResult {
	defaultMessage := fmt.Sprintf("%v does not equal %v", obj, obj2)

	if obj == nil || obj2 == nil {
		return v.addValidationResult(name, false, defaultMessage)
	}

	return v.addValidationResult(name, reflect.DeepEqual(obj, obj2), defaultMessage)
}

func (v *Validation) MatchRegex(name string, obj interface{}, pattern string) *ValidationResult {
	defaultMessage := fmt.Sprintf("Must match regex %s", pattern)

	if obj == nil {
		return v.addValidationResult(name, false, defaultMessage)
	}

  matched, err := regexp.MatchString(pattern, reflect.ValueOf(obj).String())
  if err != nil {
  	return v.addValidationResult(name, false, defaultMessage)
  }

  return v.addValidationResult(name, matched, defaultMessage)
}

func (v *Validation) Custom(name string, match bool, message string) *ValidationResult {
	return v.addValidationResult(name, match, message)
}