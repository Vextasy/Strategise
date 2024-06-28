package internal

import (
	"fmt"
	"reflect"
)

// getStructTag returns the struct tag for the given field name and tag key.
// For example:
//
//	type Snapshot struct {
//	  Date  time.Time  `format:"2006-01-02"`}
//	}
//
// getStructTag(Snapshot{}, "Date", "format") returns "2006-01-02"
func GetStructTag(t interface{}, fieldName, tagKey string) (string, error) {
	rt := reflect.TypeOf(t)
	if rt.Kind() != reflect.Struct {
		return "", fmt.Errorf("bad type")
	}

	field, found := rt.FieldByName(fieldName)
	if !found {
		return "", fmt.Errorf("field not found")
	}

	return field.Tag.Get(tagKey), nil
}
