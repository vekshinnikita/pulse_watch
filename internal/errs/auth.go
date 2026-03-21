package errs

const (
	InvalidLoginOrPasswordErrorMessage = "Invalid username or password"
	TokenExpiredErrorMessage           = "Token is expired"
	TokenRevokedErrorMessage           = "Token is revoked"
	InvalidTokenErrorMessage           = "Invalid token"
	UnauthorizedErrorMessage           = "Authorization required"
	InvalidAuthHeaderErrorMessage      = "Invalid authorization header"
	ForbiddenErrorMessage              = "Access is denied"
)

type ExpiredOrRevokedTokenError struct {
	Message string
}

func (e *ExpiredOrRevokedTokenError) Error() string {
	return e.Message
}

type InvalidTokenError struct {
	Message string
}

func (e *InvalidTokenError) Error() string {
	return e.Message
}
