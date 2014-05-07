package jantar

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"time"
)

type Validation struct {
	hasErrors bool
	errors    map[string][]string
}

type ValidationError struct {
	validation *Validation
	name       string
	index      int
}

func (v *Validation) HasErrors() bool {
	return v.hasErrors
}

func (v *Validation) addError(name string, message string) *ValidationError {
	result := &ValidationError{v, name, -1}

	v.hasErrors = true
	v.errors[name] = append(v.errors[name], message)
	result.index = len(v.errors[name]) - 1

	return result
}

func (vr *ValidationError) Message(msg string) *ValidationError {
	if vr != nil && vr.index != -1 {
		vr.validation.errors[vr.name][vr.index] = msg
	}

	return vr
}

// TODO: use type switch

func (v *Validation) Required(name string, obj interface{}) *ValidationError {
	valid := false
	defaultMessage := "Required"

	switch value := obj.(type) {
	case nil:
		valid = false
	case int:
		valid = value != 0
	case string:
		valid = len(value) > 0
	case time.Time:
		valid = value.IsZero()
	default:
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice {
			valid = v.Len() > 0
		}
	}

	if !valid {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *Validation) Min(name string, obj interface{}, min int) *ValidationError {
	valid := false
	defaultMessage := fmt.Sprintf("Must be larger than %d", min)

	switch value := obj.(type) {
	case nil:
		valid = false
	case int:
		valid = value >= min
	case string:
		valid = len(value) >= min
	default:
		v := reflect.ValueOf(obj)
		if v.Kind() == reflect.Slice {
			valid = v.Len() >= min
		}
	}

	if !valid {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *Validation) Max(name string, obj interface{}, max int) *ValidationError {
	valid := false
	defaultMessage := fmt.Sprintf("Must be smaller than %d", max)

	switch value := obj.(type) {
	case nil:
		valid = false
	case int:
		valid = value <= max
	case string:
		valid = len(value) <= max
	default:
		v := reflect.ValueOf(obj)
		if v.Kind() == reflect.Slice {
			valid = v.Len() <= max
		}
	}

	if !valid {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *Validation) MinMax(name string, obj interface{}, min int, max int) *ValidationError {
	valid := false
	defaultMessage := fmt.Sprintf("Must be larger %d and smaller %d", min, max)

	switch value := obj.(type) {
	case nil:
		valid = false
	case int:
		valid = value >= min && value <= max
	case string:
		valid = len(value) >= min && len(value) <= max
	default:
		v := reflect.ValueOf(obj)
		if v.Kind() == reflect.Slice {
			valid = v.Len() >= min && v.Len() <= max
		}
	}

	if !valid {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *Validation) Length(name string, obj interface{}, length int) *ValidationError {
	valid := false
	defaultMessage := fmt.Sprintf("Must be %d symbols long", length)

	switch value := obj.(type) {
	case nil:
		valid = false
	case int:
		valid = int(math.Ceil(math.Log10(float64(value)))) == length
	case string:
		valid = len(value) == length
	default:
		v := reflect.ValueOf(obj)
		if v.Kind() == reflect.Slice {
			valid = v.Len() == length
		}
	}

	if !valid {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *Validation) Equals(name string, obj interface{}, obj2 interface{}) *ValidationError {
	defaultMessage := fmt.Sprintf("%v does not equal %v", obj, obj2)

	if obj == nil || obj2 == nil || !reflect.DeepEqual(obj, obj2) {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *Validation) MatchRegex(name string, obj interface{}, pattern string) *ValidationError {
	valid := true
	defaultMessage := fmt.Sprintf("Must match regex %s", pattern)

	if obj == nil {
		valid = false
	} else {
		match, err := regexp.MatchString(pattern, reflect.ValueOf(obj).String())
		if err != nil || !match {
			valid = false
		}
	}

	if !valid {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *Validation) Custom(name string, match bool, message string) *ValidationError {
	if match {
		return v.addError(name, message)
	}

	return nil
}
