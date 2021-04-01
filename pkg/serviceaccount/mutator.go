package serviceaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"registry-secret-manager/pkg/registry"
	"registry-secret-manager/pkg/secret"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Mutator struct {
	client     client.Client
	registries []registry.Registry

	decoder *admission.Decoder
}

func NewMutator(client client.Client, registries []registry.Registry) *Mutator {
	return &Mutator{
		client:     client,
		registries: registries,
	}
}

func (m *Mutator) Handle(ctx context.Context, request admission.Request) admission.Response {
	log.Debugf("Received request to mutate ServiceAccount [%s/%s]", request.Namespace, request.Name)

	// Decode the ServiceAccount from the request
	serviceAccount := &corev1.ServiceAccount{}

	err := m.decoder.Decode(request, serviceAccount)
	if err != nil {
		err = fmt.Errorf("failed to decode ServiceAccount [%s/%s]: %w", request.Namespace, request.Namespace, err)
		log.Error(err)

		return admission.Errored(http.StatusBadRequest, err)
	}

	// Create the secret if needed
	err = secret.CreateSecretIfNeeded(ctx, m.client, m.registries, request.Namespace)
	if err != nil {
		// We should not prevent the ServiceAccount from being mutated if the Secret creation fails.
		// This is safe to do as the Reconciler will attempt to create the Secret anyway.
		err := fmt.Errorf("failed to create the secret, but ignoring the error: %w", err)
		log.Error(err)

		return admission.Errored(http.StatusFailedDependency, err)
	}

	// Mutate the ServiceAccount if needed
	if !needsMutation(serviceAccount) {
		reason := fmt.Sprintf("No mutation needed for ServiceAccount [%s/%s]", request.Namespace, request.Name)
		log.Debug(reason)

		return admission.Allowed(reason)
	}

	// Patch the ServiceAccount with the secret
	log.Infof("Responding with a patch to ServiceAccount [%s/%s]", request.Namespace, request.Name)

	serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{
		Name: "registry-secret",
	})

	patched, err := json.Marshal(serviceAccount)
	if err != nil {
		err = fmt.Errorf("failed to marshall ServiceAccount [%s/%s]: %w", request.Namespace, request.Namespace, err)
		log.Error(err)

		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(request.Object.Raw, patched)
}

func (m *Mutator) InjectDecoder(decoder *admission.Decoder) (err error) {
	m.decoder = decoder

	return
}
