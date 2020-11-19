package registry

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	cred "github.com/aws/aws-sdk-go/aws/credentials"
	sess "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type ECR struct {
}

func NewECR() *ECR {
	return &ECR{}
}

func (e *ECR) Login() (*Credentials, error) {
	config := aws.NewConfig()
	config.WithRegion(os.Getenv("AWS_DEFAULT_REGION"))
	config.WithCredentials(cred.NewEnvCredentials())

	session, err := sess.NewSessionWithOptions(sess.Options{
		Config: *config,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %v", err)
	}

	service := ecr.New(session)
	input := &ecr.GetAuthorizationTokenInput{}

	token, err := service.GetAuthorizationToken(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization token: %v", err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(*token.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode the token: %v", err)
	}

	parts := strings.Split(string(decodedBytes), ":")

	return NewCredentials(parts[0], parts[1], *token.AuthorizationData[0].ProxyEndpoint), nil
}
