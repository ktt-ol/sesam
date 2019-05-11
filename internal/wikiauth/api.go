package wikiauth

type AuthError struct {
	Error         error
	LoginNotFound bool
	SystemError   bool
}

type WikiAuth interface {
	CheckPassword(emailOrName string, password string) (userName string, errResult *AuthError)
}
