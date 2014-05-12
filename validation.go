package jantar

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"time"
)

/* TODO: allow custom cookie name */

// Validation is a helper for validating user supplied data and returning error messages.
// It offers various validation functions and can save errors in a Cookie.
type Validation struct {
	rw        http.ResponseWriter
	hasErrors bool
	errors    map[string][]string
}

type validationError struct {
	validation *Validation
	name       string
	index      int
}

func newValidation(rw http.ResponseWriter) *Validation {
	return &Validation{rw, false, make(map[string][]string)}
}

// SaveErrors saves current validation error in a http.Cookie
func (v *Validation) SaveErrors() {
	if v.hasErrors {
		values := url.Values{}
		for key, array := range v.errors {
			for _, val := range array {
				values.Add(key, val)
			}
		}

		http.SetCookie(v.rw, &http.Cookie{Name: "JANTAR_ERRORS", Value: values.Encode(), Secure: false, HttpOnly: true, Path: "/"})
	}
}

// HasErrors returns true of an validation error occured. Otherwise false is returned
func (v *Validation) HasErrors() bool {
	return v.hasErrors
}

func (v *Validation) addError(name string, message string) *validationError {
	result := &validationError{v, name, -1}

	v.hasErrors = true
	v.errors[name] = append(v.errors[name], message)
	result.index = len(v.errors[name]) - 1

	return result
}

// Required checks the existance of given obj. How exactly this check is being performed depends on the type of obj. Valid types are: int, string, time.Time and slice.
// The given name determines the association of this error in the resulting validation error map.
func (v *Validation) Required(name string, obj interface{}) *validationError {
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

// TODO: add time.Time to Min

// Min checks if given obj is smaller or equal to min. How exactly this check is being performed depends on the type of obj. Valid types are: int, string and slice.
// The given name determines the association of this error in the resulting validation error map.
func (v *Validation) Min(name string, obj interface{}, min int) *validationError {
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

// TODO: add time.Time to Max

// Max checks if given obj is smaller or equal max. How exactly this check is being performed depends on the type of obj. Valid types are: int, string and slice.
// The given name determines the association of this error in the resulting validation error map.
func (v *Validation) Max(name string, obj interface{}, max int) *validationError {
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

// TODO: add time.Time to MinMax

// MinMax compiles Min and Max in one call.
func (v *Validation) MinMax(name string, obj interface{}, min int, max int) *validationError {
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

// Length checks the exact length of obj. How exactly this check is being performed depends on the type of obj. Valid types are: int, string and slice.
// The given name determines the association of this error in the resulting validation error map.
func (v *Validation) Length(name string, obj interface{}, length int) *validationError {
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

// Equal tests for deep equality of two given objects.
// The given name determines the association of this error in the resulting validation error map.
func (v *Validation) Equals(name string, obj interface{}, obj2 interface{}) *validationError {
	defaultMessage := fmt.Sprintf("%v does not equal %v", obj, obj2)

	if obj == nil || obj2 == nil || !reflect.DeepEqual(obj, obj2) {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *Validation) MatchRegex(name string, obj interface{}, pattern string) *validationError {
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

func (v *Validation) Custom(name string, match bool, message string) *validationError {
	if match {
		return v.addError(name, message)
	}

	return nil
}

func (vr *validationError) Message(msg string) *validationError {
	if vr != nil && vr.index != -1 {
		vr.validation.errors[vr.name][vr.index] = msg
	}

	return vr
}
