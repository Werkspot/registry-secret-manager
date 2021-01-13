package secret

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcile(t *testing.T) {
	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "registry-secret-manager",
			Name:      "registry-secret",
		},
	}

	secret := &corev1.Secret{
		TypeMeta: SecretTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "registry-secret-manager",
			Name:            "registry-secret",
			ResourceVersion: "1",
		},
	}

	fakeClient := fake.NewFakeClient(secret)
	reconciler := newReconciler(fakeClient, nil)

	// Reconcile and verify its content
	result, err := reconciler.Reconcile(context.TODO(), request)

	assert.NoError(t, err)
	assert.True(t, result.Requeue)
	assert.Equal(t, 3*time.Hour, result.RequeueAfter)

	// Retrieve and verify if the Secret was updated
	secret = &corev1.Secret{}
	err = fakeClient.Get(context.TODO(), request.NamespacedName, secret)

	assert.NoError(t, err)
	assert.Equal(t, "2", secret.ResourceVersion)
	assert.Equal(t, corev1.SecretTypeDockerConfigJson, secret.Type)
}
