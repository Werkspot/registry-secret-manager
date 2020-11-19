package registry

type Credentials struct {
	Username string
	Password string
	Endpoint string
}

func NewCredentials(username, password, endpoint string) *Credentials {
	return &Credentials{
		Username: username,
		Password: password,
		Endpoint: endpoint,
	}
}
