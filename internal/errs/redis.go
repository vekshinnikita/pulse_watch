package errs

type NotExistKeyRedisError struct {
	Message string
}

func (e *NotExistKeyRedisError) Error() string {
	return e.Message
}
