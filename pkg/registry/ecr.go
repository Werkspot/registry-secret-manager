package registry

import (
	"encoding/base64"
	"fmt"
	"strings"

	sess "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

// EcrName contains a unique name.
const EcrName = "ecr"

// ECR represents an ECR Registry.
type ECR struct{}

// NewECR returns a pointer to ECR.
func NewECR() *ECR {
	return &ECR{}
}

// Login returns a valid Credentials pointer and/or error.
func (e *ECR) Login() (*Credentials, error) {
	session, err := sess.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	service := ecr.New(session)
	input := &ecr.GetAuthorizationTokenInput{}

	token, err := service.GetAuthorizationToken(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization token: %w", err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*token.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode the token: %w", err)
	}

	parts := strings.Split(string(decodedBytes), ":")

	return NewCredentials(parts[0], parts[1], *token.AuthorizationData[0].ProxyEndpoint), nil
}
