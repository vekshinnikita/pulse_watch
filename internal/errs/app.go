package errs

const (
	AppNotFoundErrorMessage            = "app not found"
	ApiKeyNotFoundErrorMessage         = "api key not found"
	InvalidApiKeyErrorMessage          = "invalid api key"
	ExpiredOrRevokedApiKeyErrorMessage = "expired or revoked api key"
)

type InvalidApiKeyError struct {
	Message string
}

func (e *InvalidApiKeyError) Error() string {
	return e.Message
}

type ExpiredOrRevokedApiKeyError struct {
	Message string
}

func (e *ExpiredOrRevokedApiKeyError) Error() string {
	return e.Message
}
