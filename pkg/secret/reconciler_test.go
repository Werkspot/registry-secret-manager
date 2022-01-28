package secret_test

import (
	"context"
	"registry-secret-manager/pkg/secret"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcile(t *testing.T) {
	t.Parallel()

	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "registry-secret-manager",
			Name:      "registry-secret",
		},
	}

	secretObject := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.Version,
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "registry-secret-manager",
			Name:            "registry-secret",
			ResourceVersion: "1",
		},
	}

	fakeClientBuilder := fake.NewClientBuilder()
	fakeClientBuilder.WithObjects(secretObject)

	fakeClient := fakeClientBuilder.Build()
	reconciler := secret.NewReconciler(fakeClient, nil)

	// Reconcile and verify its content
	result, err := reconciler.Reconcile(context.TODO(), request)

	assert.NoError(t, err)
	assert.True(t, result.Requeue)
	assert.Equal(t, secret.ReconcileAfter, result.RequeueAfter)

	// Retrieve and verify if the Secret was updated
	secretObject = &corev1.Secret{}
	err = fakeClient.Get(context.TODO(), request.NamespacedName, secretObject)

	assert.NoError(t, err)
	assert.Equal(t, "2", secretObject.ResourceVersion)
	assert.Equal(t, corev1.SecretTypeDockerConfigJson, secretObject.Type)
}
