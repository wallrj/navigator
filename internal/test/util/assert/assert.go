package assert

import (
	"reflect"
	"strings"
	"testing"
)

func ObjectMatches(t *testing.T, actual interface{}, expected map[string]interface{}) {
	// Get to the actual type of the supplied object
	rv := reflect.ValueOf(actual)
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	for key, expectedVal := range expected {
		v := rv
		parts := strings.Split(key, ".")
		for _, part := range parts {
			v = v.FieldByName(part)
			if !v.IsValid() {
				t.Fatalf("Attribute %q not found while checking for part %q", key, part)
			}
		}
		attr := v.Interface()
		if !reflect.DeepEqual(expectedVal, attr) {
			t.Errorf("%s %#v != %#v", key, expectedVal, attr)
		}
	}
}
