package serviceaccount

import (
	"fmt"
	"registry-secret-manager/pkg/registry"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// NewController initializes a service account controller.
func NewController(mgr manager.Manager, registries []registry.Registry) error {
	// Setup the webhooks
	server := mgr.GetWebhookServer()
	server.Register("/mutate", &webhook.Admission{
		Handler: NewMutator(mgr.GetClient(), registries),
	})

	// Setup the reconciler
	serviceAccountController, err := controller.New("serviceaccount", mgr, controller.Options{
		Reconciler: NewReconciler(mgr.GetClient(), registries),
	})
	if err != nil {
		return fmt.Errorf("unable to set up ServiceAccount controller: %w", err)
	}

	// Watch ServiceAccounts and enqueue ServiceAccount object key
	err = serviceAccountController.Watch(
		&source.Kind{
			Type: &corev1.ServiceAccount{},
		},
		&handler.EnqueueRequestForObject{},
		predicate.Funcs{
			DeleteFunc: func(event event.DeleteEvent) bool {
				log.Debugf(
					"Skipping reconciliation of ServiceAccount [%s/%s] as it has been deleted",
					event.Object.GetNamespace(),
					event.Object.GetName(),
				)

				return false
			},
			GenericFunc: func(event event.GenericEvent) bool {
				log.Debugf(
					"Skipping reconciliation of ServiceAccount [%s/%s] for the generic event type",
					event.Object.GetNamespace(),
					event.Object.GetName(),
				)

				return false
			},
		},
	)
	if err != nil {
		return fmt.Errorf("unable to watch ServiceAccounts: %w", err)
	}

	return nil
}
