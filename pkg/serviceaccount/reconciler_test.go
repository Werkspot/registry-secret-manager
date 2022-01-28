package serviceaccount_test

import (
	"context"
	"registry-secret-manager/pkg/serviceaccount"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestReconcile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		mustCreateSecret bool
		existing         *corev1.ServiceAccount
		expected         *corev1.ServiceAccount
	}{
		{
			name:             "no patch needed, must create the secret",
			mustCreateSecret: true,
			existing:         newServiceAccount(1, "registry-secret"),
			expected:         newServiceAccount(1, "registry-secret"),
		},
		{
			name:             "no secrets at all, must create the secret",
			mustCreateSecret: true,
			existing:         newServiceAccount(1),
			expected:         newServiceAccount(2, "registry-secret"),
		},
		{
			name:             "no secrets managed by us, must create the secret",
			mustCreateSecret: true,
			existing:         newServiceAccount(1, "not-managed-by-us"),
			expected:         newServiceAccount(2, "not-managed-by-us", "registry-secret"),
		},

		{
			name:             "no patch needed, must not create the secret",
			mustCreateSecret: false,
			existing:         newServiceAccount(1, "registry-secret"),
			expected:         newServiceAccount(1, "registry-secret"),
		},
		{
			name:             "no secrets at all, must not create the secret",
			mustCreateSecret: false,
			existing:         newServiceAccount(1),
			expected:         newServiceAccount(2, "registry-secret"),
		},
		{
			name:             "no secrets managed by us, must not create the secret",
			mustCreateSecret: false,
			existing:         newServiceAccount(1, "not-managed-by-us"),
			expected:         newServiceAccount(2, "not-managed-by-us", "registry-secret"),
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertReconcile(t, test.existing, test.expected, test.mustCreateSecret)
		})
	}
}

func assertReconcile(t *testing.T, existing, expected *corev1.ServiceAccount, mustCreateTheSecret bool) {
	t.Helper()

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
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.Version,
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace:       secretName.Namespace,
				Name:            secretName.Name,
				ResourceVersion: "1",
			},
		})
	}

	fakeClientBuilder := fake.NewClientBuilder()
	fakeClientBuilder.WithObjects(objects...)

	fakeClient := fakeClientBuilder.Build()
	reconciler := serviceaccount.NewReconciler(fakeClient, nil)

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
