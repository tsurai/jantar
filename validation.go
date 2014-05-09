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

type validation struct {
	rw        http.ResponseWriter
	hasErrors bool
	errors    map[string][]string
}

type validationError struct {
	validation *validation
	name       string
	index      int
}

func newvalidation(rw http.ResponseWriter) *validation {
	return &validation{rw, false, make(map[string][]string)}
}

func (v *validation) SaveErrors() {
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

func (v *validation) HasErrors() bool {
	return v.hasErrors
}

func (v *validation) addError(name string, message string) *validationError {
	result := &validationError{v, name, -1}

	v.hasErrors = true
	v.errors[name] = append(v.errors[name], message)
	result.index = len(v.errors[name]) - 1

	return result
}

func (vr *validationError) Message(msg string) *validationError {
	if vr != nil && vr.index != -1 {
		vr.validation.errors[vr.name][vr.index] = msg
	}

	return vr
}

func (v *validation) Required(name string, obj interface{}) *validationError {
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

func (v *validation) Min(name string, obj interface{}, min int) *validationError {
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

func (v *validation) Max(name string, obj interface{}, max int) *validationError {
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

func (v *validation) MinMax(name string, obj interface{}, min int, max int) *validationError {
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

func (v *validation) Length(name string, obj interface{}, length int) *validationError {
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

func (v *validation) Equals(name string, obj interface{}, obj2 interface{}) *validationError {
	defaultMessage := fmt.Sprintf("%v does not equal %v", obj, obj2)

	if obj == nil || obj2 == nil || !reflect.DeepEqual(obj, obj2) {
		return v.addError(name, defaultMessage)
	}

	return nil
}

func (v *validation) MatchRegex(name string, obj interface{}, pattern string) *validationError {
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

func (v *validation) Custom(name string, match bool, message string) *validationError {
	if match {
		return v.addError(name, message)
	}

	return nil
}
