package secret

import (
	"encoding/json"
	"testing"

	"registry-secret-manager/pkg/registry"

	"github.com/stretchr/testify/assert"
)

func TestNewDockerConfig(t *testing.T) {
	tests := map[string]struct {
		credentials []*registry.Credentials
		expected    string
	}{
		"empty credentials": {
			credentials: nil,
			expected:    `{"auths":{}}`,
		},
		"one credential": {
			credentials: []*registry.Credentials{
				{
					Username: "user",
					Password: "pass",
					Endpoint: "https://foo.bar",
				},
			},
			expected: `{"auths":{"https://foo.bar":{"username":"user","password":"pass","email":"` + defaultEmail + `","auth":"dXNlcjpwYXNz"}}}`,
		},
		"more than one credential": {
			credentials: []*registry.Credentials{
				{
					Username: "one",
					Password: "pass",
					Endpoint: "https://foo.bar/one",
				},
				{
					Username: "two",
					Password: "pass",
					Endpoint: "https://foo.bar/two",
				},
			},
			expected: `{"auths":{` +
				`"https://foo.bar/one":{"username":"one","password":"pass","email":"` + defaultEmail + `","auth":"b25lOnBhc3M="},` +
				`"https://foo.bar/two":{"username":"two","password":"pass","email":"` + defaultEmail + `","auth":"dHdvOnBhc3M="}}}`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			secret := NewDockerConfig(test.credentials)
			result, err := json.Marshal(secret)

			assert.NoError(t, err)
			assert.Equal(t, string(result), test.expected)
		})
	}
}
