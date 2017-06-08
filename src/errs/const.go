package errs

// ConstantError is a constant error on string
type ConstantError string

func (c ConstantError) Error() string {
	return string(c)
}

const (
	// InternalServerError is a message for 500 status
	InternalServerError = "something went wrong"

	// ModelNotFound in store
	ModelNotFound ConstantError = "model not found"
)
