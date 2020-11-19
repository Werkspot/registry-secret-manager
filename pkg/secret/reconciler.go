package secret

import (
	"context"
	"fmt"
	"time"

	"registry-secret-manager/pkg/registry"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reconciler struct {
	client     client.Client
	registries []registry.Registry
}

func newReconciler(client client.Client, registries []registry.Registry) *reconciler {
	return &reconciler{
		client:     client,
		registries: registries,
	}
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log.Debugf("Received request to reconcile Secret [%s]", request.NamespacedName)

	// We requeue the reconciliation so that we keep on renewing the authorization lifetime (eg: for ECR)
	result := reconcile.Result{
		Requeue:      true,
		RequeueAfter: 3 * time.Hour,
	}

	// Fetch the Secret from cache
	secret := &corev1.Secret{}

	err := r.client.Get(ctx, request.NamespacedName, secret)
	if errors.IsNotFound(err) {
		log.Debugf("Stopping reconciliation of Secret [%s] as it no longer exists: %v", request.NamespacedName, err)
		return reconcile.Result{}, nil
	}
	if err != nil {
		err = fmt.Errorf("could not fetch the Secret [%s]: %v", request.NamespacedName, err)
		log.Error(err)
		return result, err
	}

	// Update the Secret
	secret, err = createSecretObject(r.registries, request.Namespace)
	if err != nil {
		err = fmt.Errorf("could not create the Secret object [%s]: %v", request.NamespacedName, err)
		log.Error(err)
		return result, err
	}

	err = r.client.Update(ctx, secret)
	if err != nil {
		err = fmt.Errorf("could not update the Secret [%s]: %v", request.NamespacedName, err)
		log.Error(err)
		return result, err
	}

	log.Infof("Sucessfully updated the Secret [%s]", request.NamespacedName)

	return result, nil
}
