package errs

const (
	UserNotFoundErrorMessage = "user not found"
)

type UserNotFoundError struct {
	Message string
}

func (e *UserNotFoundError) Error() string {
	return e.Message
}
