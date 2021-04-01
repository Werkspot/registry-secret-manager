package registry

// Credentials represents a credentials object.
type Credentials struct {
	Username string
	Password string
	Endpoint string
}

// NewCredentials returns a pointer to Credentials.
func NewCredentials(username, password, endpoint string) *Credentials {
	return &Credentials{
		Username: username,
		Password: password,
		Endpoint: endpoint,
	}
}
