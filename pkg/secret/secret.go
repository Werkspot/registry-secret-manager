package secret

import (
	"context"
	"encoding/json"
	"fmt"

	reg "registry-secret-manager/pkg/registry"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SecretTypeMeta containts the TypeMeta for a secret
var SecretTypeMeta = metav1.TypeMeta{
	APIVersion: corev1.SchemeGroupVersion.Version,
	Kind:       "Secret",
}

// CreateSecretIfNeeded on the given namespace if it doesn't already exist
func CreateSecretIfNeeded(ctx context.Context, client client.Client, registries []reg.Registry, namespace string) error {
	secretName := types.NamespacedName{
		Namespace: namespace,
		Name:      "registry-secret",
	}
	secret := &corev1.Secret{}

	err := client.Get(ctx, secretName, secret)
	if err == nil {
		log.Debugf("No need to create the already existing Secret [%s]", secretName)
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("could not fetch the Secret [%s]: %v", secretName, err)
	}

	// Secret is not found, we create it now
	secret, err = createSecretObject(registries, namespace)
	if err != nil {
		return fmt.Errorf("failed to create Secret [%s]: %v", secretName, err)
	}

	err = client.Create(ctx, secret)
	if err == nil {
		log.Infof("Sucessfully created the Secret [%s]", secretName)
		return nil
	}
	if errors.IsAlreadyExists(err) {
		// Because creating the object and the secret can take some time it is possible that another process already
		// created the desired Secret. We can safely ignore the error.
		log.Debugf("No need to create the already existing Secret [%s]", secretName)
		return nil
	}

	return fmt.Errorf("could not create Secret [%s]: %v", secretName, err)
}

func createSecretObject(registries []reg.Registry, namespace string) (*corev1.Secret, error) {
	var registryCredentials []*reg.Credentials
	for _, registry := range registries {
		credentials, err := registry.Login()
		if err != nil {
			return nil, fmt.Errorf("failed to login: %v", err)
		}

		registryCredentials = append(registryCredentials, credentials)
	}

	dockerConfig := NewDockerConfig(registryCredentials)
	dockerConfigBytes, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall json: %v", err)
	}

	secret := &corev1.Secret{
		TypeMeta: SecretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "registry-secret",
			Labels: map[string]string{
				"app.kubernetes.io/name": "registry-secret-manager",
				"registry-secret":        "true",
			},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		StringData: map[string]string{
			corev1.DockerConfigJsonKey: string(dockerConfigBytes),
		},
	}

	return secret, nil
}
