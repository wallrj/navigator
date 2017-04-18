package errors

type transientInterface interface {
	error
	transient() bool
}

type transientError struct { error }

func (t transientError) transient() bool {
	return true
}

func IsTransient(err error) bool {
	_, ok := err.(transientError)
	return ok
}

func Transient(err error) error {
	return transientError{err}
}
