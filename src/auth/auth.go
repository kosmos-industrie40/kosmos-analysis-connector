package auth

// Auth is the interface with will be used to provide the authentication mechanism
type Auth interface {
	Login() error
	Logout() error
}

// Token contains the token, which will be used to authenticate
type Token interface {
	// return the token as string
	AuthToken() string
}
