package serviceaccount

import (
	"context"
	"fmt"
	"registry-secret-manager/pkg/registry"
	"registry-secret-manager/pkg/secret"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client     client.Client
	registries []registry.Registry
}

func NewReconciler(client client.Client, registries []registry.Registry) *Reconciler {
	return &Reconciler{
		client:     client,
		registries: registries,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Debugf("Received request to reconcile ServiceAccount [%s]", request.NamespacedName)

	// Fetch the ServiceAccount from cache
	result := reconcile.Result{}
	serviceAccount := &corev1.ServiceAccount{}

	err := r.client.Get(ctx, request.NamespacedName, serviceAccount)
	if errors.IsNotFound(err) {
		log.Debugf("Stopping reconciliation of ServiceAccount [%s] as it no longer exists: %v", request.NamespacedName, err)

		return reconcile.Result{}, nil
	}

	if err != nil {
		err = fmt.Errorf("could not fetch the ServiceAccount [%s]: %w", request.NamespacedName, err)
		log.Error(err)

		return result, err
	}

	// Create the secret if needed
	err = secret.CreateSecretIfNeeded(ctx, r.client, r.registries, request.Namespace)
	if err != nil {
		err = fmt.Errorf("%w", err)
		log.Error(err)

		return result, err
	}

	// Mutate the ServiceAccount if needed
	if !needsMutation(serviceAccount) {
		log.Debugf("No reconcile needed for ServiceAccount [%s]", request.NamespacedName)

		return result, nil
	}

	serviceAccount.ImagePullSecrets = append(serviceAccount.ImagePullSecrets, corev1.LocalObjectReference{
		Name: "registry-secret",
	})

	err = r.client.Update(ctx, serviceAccount)
	if err != nil {
		err = fmt.Errorf("could not update ServiceAccount [%s]: %w", request.NamespacedName, err)
		log.Error(err)

		return result, err
	}

	log.Infof("Successfully updated the ServiceAccount [%s]", request.NamespacedName)

	return result, nil
}
