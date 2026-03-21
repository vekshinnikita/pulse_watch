package errs

const (
	InvalidInputErrorMessage string = "Invalid input"
	InternalErrorMessage     string = "Internal Server Error"
)

type CheckViolationError struct {
	Message string
}

func (e *CheckViolationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}
