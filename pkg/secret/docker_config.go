package secret

import (
	"encoding/base64"
	"fmt"

	"registry-secret-manager/pkg/registry"
)

// Having an email set in the configuration is mandatory and without it the authentication will fail.
// However, any valid email can be used.
// See: https://github.com/kubernetes/kubernetes/issues/41727
var defaultEmail = "technology@werkspot.nl"

// DockerConfig stores a map of valid Authorization
type DockerConfig struct {
	Authorizations map[string]Authorization `json:"auths"`
}

// Authorization contains a valid set of credentials
type Authorization struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Auth     string `json:"auth"`
}

// NewDockerConfig returns a pointer to DockerConfig
func NewDockerConfig(registryCredentials []*registry.Credentials) *DockerConfig {
	authorizations := map[string]Authorization{}

	for _, credentials := range registryCredentials {
		token := fmt.Sprintf("%s:%s", credentials.Username, credentials.Password)
		tokenBytes := []byte(token)

		authorizations[credentials.Endpoint] = Authorization{
			Username: credentials.Username,
			Password: credentials.Password,
			Email:    defaultEmail,
			Auth:     base64.StdEncoding.EncodeToString(tokenBytes),
		}
	}

	return &DockerConfig{
		Authorizations: authorizations,
	}
}
