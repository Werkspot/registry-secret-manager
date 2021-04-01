package secret_test

import (
	"encoding/json"
	"registry-secret-manager/pkg/registry"
	"registry-secret-manager/pkg/secret"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDockerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		credentials []*registry.Credentials
		expected    string
	}{
		{
			name:        "empty credentials",
			credentials: nil,
			expected:    `{"auths":{}}`,
		},
		{
			name: "one credential",
			credentials: []*registry.Credentials{
				{
					Username: "user",
					Password: "pass",
					Endpoint: "https://foo.bar",
				},
			},
			expected: `{"auths":{"https://foo.bar":{"username":"user","password":"pass","email":"` + secret.DefaultEmail + `","auth":"dXNlcjpwYXNz"}}}`,
		},
		{
			name: "more than one credential",
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
				`"https://foo.bar/one":{"username":"one","password":"pass","email":"` + secret.DefaultEmail + `","auth":"b25lOnBhc3M="},` +
				`"https://foo.bar/two":{"username":"two","password":"pass","email":"` + secret.DefaultEmail + `","auth":"dHdvOnBhc3M="}}}`,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			s := secret.NewDockerConfig(test.credentials)
			result, err := json.Marshal(s)

			assert.NoError(t, err)
			assert.Equal(t, string(result), test.expected)
		})
	}
}
