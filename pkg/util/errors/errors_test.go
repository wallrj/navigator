package errors

import (
	"testing"
	"errors"
)

func TestIsTransient(t *testing.T) {
	type testDef struct {
		Name string
		Err error
		IsTransient bool
	}

	tests := []testDef{
		{
			Name: "test is transient",
			Err: transientError{errors.New("transient err")},
			IsTransient: true,
		},
		{
			Name: "test is not transient",
			Err: errors.New("not transient err"),
			IsTransient: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(test testDef) func(t *testing.T) {
			return func(t *testing.T) {
				if it := IsTransient(test.Err); it != test.IsTransient {
					t.Error("expected IsTransient to return '%v', but was '%v'", test.IsTransient, it)
				}
			}
		}(test))
	}

	//	: errors.New("err"),

}