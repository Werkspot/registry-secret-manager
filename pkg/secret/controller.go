package secret

import (
	"fmt"

	"registry-secret-manager/pkg/registry"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NewController initializes a secret controller
func NewController(mgr manager.Manager, registries []registry.Registry) error {
	// Setup the reconciler
	secretController, err := controller.New("secret", mgr, controller.Options{
		Reconciler: newReconciler(mgr.GetClient(), registries),
	})
	if err != nil {
		return fmt.Errorf("unable to set up Secret controller: %v", err)
	}

	// Only handle the Secrets that matches these labels
	labelSelector, err := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app.kubernetes.io/name": "registry-secret-manager",
			"registry-secret":        "true",
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create label selector for Secrets: %v", err)
	}

	// Watch Secrets and enqueue Secret object key
	err = secretController.Watch(
		&source.Kind{
			Type: &corev1.Secret{},
		},
		&handler.EnqueueRequestForObject{},
		labelSelector,
		predicate.Funcs{
			// Skip everything but the create event, we want to have an initial reconciliation (create event), and keep
			// on periodically reconciling. But we don't need to reconcile update as it already contains the correct/desired
			// registry credentials. Otherwise it will end up in a loop.
			// This could be improved, however, by checking the contents of the secret and if its due for renewing.
			UpdateFunc: func(event event.UpdateEvent) bool {
				log.Debugf(
					"Skipping reconciliation of Secret [%s/%s] as it has just been updated",
					event.ObjectNew.GetNamespace(),
					event.ObjectNew.GetName(),
				)
				return false
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				log.Debugf(
					"Skipping reconciliation of Secret [%s/%s] as it has been deleted",
					event.Object.GetNamespace(),
					event.Object.GetName(),
				)
				return false
			},
			GenericFunc: func(event event.GenericEvent) bool {
				log.Debugf(
					"Skipping reconciliation of Secret [%s/%s] for the generic event type",
					event.Object.GetNamespace(),
					event.Object.GetName(),
				)
				return false
			},
		},
	)
	if err != nil {
		return fmt.Errorf("unable to watch Secrets: %v", err)
	}

	return nil
}
