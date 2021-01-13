package registry

import (
	"fmt"
	"os"
)

// DockerHubName contains a unique name
const DockerHubName = "docker-hub"

// DockerHub represents a Docker Hub Registry
type DockerHub struct {
}

// NewDockerHub returns a pointer to DockerHub
func NewDockerHub() *DockerHub {
	return &DockerHub{}
}

// Login returns a valid Credentials pointer and/or error
func (d *DockerHub) Login() (*Credentials, error) {
	username, err := d.retrieveEnvVar("DOCKER_HUB_USERNAME")
	if err != nil {
		return nil, err
	}

	password, err := d.retrieveEnvVar("DOCKER_HUB_PASSWORD")
	if err != nil {
		return nil, err
	}

	endpoint, err := d.retrieveEnvVar("DOCKER_HUB_ENDPOINT")
	if err != nil {
		return nil, err
	}

	return NewCredentials(username, password, endpoint), nil
}

func (d *DockerHub) retrieveEnvVar(key string) (string, error) {
	value, present := os.LookupEnv(key)
	if !present {
		return value, fmt.Errorf("could not find environment value for %s", key)
	}
	if value == "" {
		return value, fmt.Errorf("found empty environment value for %s", key)
	}
	return value, nil
}
