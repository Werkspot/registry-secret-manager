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

func NewController(mgr manager.Manager, registries []registry.Registry) error {
	// Setup the webhooks
	server := mgr.GetWebhookServer()
	server.Register("/mutate", &webhook.Admission{
		Handler: newMutator(mgr.GetClient(), registries),
	})

	// Setup the reconciler
	serviceAccountController, err := controller.New("serviceaccount", mgr, controller.Options{
		Reconciler: newReconciler(mgr.GetClient(), registries),
	})
	if err != nil {
		return fmt.Errorf("unable to set up ServiceAccount controller: %v", err)
	}

	// Watch ServiceAccounts and enqueue ServiceAccount object key
	err = serviceAccountController.Watch(
		&source.Kind{
			Type: &corev1.ServiceAccount{},
		},
		&handler.EnqueueRequestForObject{},
		predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				// Skip this event when the generation hasn't changed
				if event.ObjectOld.GetGeneration() == event.ObjectNew.GetGeneration() {
					log.Debugf(
						"Skipping reconciliation of ServiceAccount [%s/%s] as it hasn't changed",
						event.ObjectNew.GetNamespace(),
						event.ObjectNew.GetName(),
					)
					return false
				}
				return true
			},
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
		return fmt.Errorf("unable to watch ServiceAccounts: %v", err)
	}

	return nil
}
