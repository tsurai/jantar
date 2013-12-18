package amber

import (
	"net/url"
	"reflect"
	"strings"
)

/*
type User struct {
	Name			string
	Email			string
	Password	string
}
*/

func ParsePostData(src url.Values, dst reflect.Type) reflect.Value {
	if len(src) <= 0 {
		logger.Println("Failed to parse post data. Data is nil")
		return reflect.ValueOf(nil)
	}

	typeName := dst.String()
	if !strings.HasPrefix(typeName, "*") {
		logger.Println("-[Warning]- " + typeName + " is no pointer in ParsePostData")
		return reflect.ValueOf(nil)
	}

	dstval := reflect.New(dst.Elem())
	name := typeName[strings.LastIndex(typeName, ".")+1:]

	for key, value := range src {
		substr := strings.SplitN(key, ".", 2)
		if (len(substr) == 2) && (strings.EqualFold(substr[0], name)) {
			dstval.Elem().FieldByName(substr[1]).Set(reflect.ValueOf(value[0]))
		}
	}

	return dstval
}