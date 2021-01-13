package serviceaccount

import (
	"context"
	"strconv"
	"testing"

	"registry-secret-manager/pkg/secret"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcile(t *testing.T) {
	tests := map[string]struct {
		existing *corev1.ServiceAccount
		expected *corev1.ServiceAccount
	}{
		"no patch needed": {
			existing: newServiceAccount(1, "registry-secret"),
			expected: newServiceAccount(1, "registry-secret"),
		},
		"no secrets at all": {
			existing: newServiceAccount(1),
			expected: newServiceAccount(2, "registry-secret"),
		},
		"no secrets managed by us": {
			existing: newServiceAccount(1, "not-managed-by-us"),
			expected: newServiceAccount(2, "not-managed-by-us", "registry-secret"),
		},
	}

	for name, test := range tests {
		t.Run(name+", must create the secret", func(t *testing.T) {
			assertReconcile(t, test.existing, test.expected, true)
		})

		t.Run(name+", must not create the secret", func(t *testing.T) {
			assertReconcile(t, test.existing, test.expected, false)
		})
	}
}

func assertReconcile(t *testing.T, existing, expected *corev1.ServiceAccount, mustCreateTheSecret bool) {
	secretName := types.NamespacedName{
		Namespace: "registry-secret-manager",
		Name:      "registry-secret",
	}

	objects := []client.Object{
		existing,
	}

	if !mustCreateTheSecret {
		// Secret is not created by the mutator
		objects = append(objects, &corev1.Secret{
			TypeMeta: secret.SecretTypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       secretName.Namespace,
				Name:            secretName.Name,
				ResourceVersion: "1",
			},
		})
	}

	// Create a client and the reconciler
	fakeClient := fake.NewFakeClient(objects...)
	reconciler := newReconciler(fakeClient, nil)

	// Reconcile and verify its content
	request := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: existing.Namespace,
			Name:      existing.Name,
		},
	}
	result, err := reconciler.Reconcile(context.TODO(), request)

	assert.NoError(t, err)
	assert.True(t, result.IsZero())

	// Retrieve and verify the contents of the reconciled ServiceAccount
	updated := &corev1.ServiceAccount{}
	err = fakeClient.Get(context.TODO(), request.NamespacedName, updated)

	assert.NoError(t, err)
	assert.Equal(t, expected.ObjectMeta, updated.ObjectMeta)
	assert.Equal(t, expected.ImagePullSecrets, updated.ImagePullSecrets)

	// Check if the secret was created or updated, either way ResourceVersion should always be 1
	var desiredSecret corev1.Secret
	err = fakeClient.Get(context.TODO(), secretName, &desiredSecret)

	assert.NoError(t, err)
	assert.Equal(t, "1", desiredSecret.ResourceVersion)
}

func newServiceAccount(resourceVersion int, imagePullSecrets ...string) *corev1.ServiceAccount {
	var secrets []corev1.LocalObjectReference
	for _, secretName := range imagePullSecrets {
		secrets = append(secrets, corev1.LocalObjectReference{
			Name: secretName,
		})
	}

	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       "registry-secret-manager",
			Name:            "default",
			ResourceVersion: strconv.Itoa(resourceVersion),
		},
		ImagePullSecrets: secrets,
	}
}
