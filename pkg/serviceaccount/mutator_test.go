package serviceaccount_test

import (
	"context"
	"encoding/json"
	"registry-secret-manager/pkg/serviceaccount"
	"testing"

	"github.com/stretchr/testify/assert"
	"gomodules.xyz/jsonpatch/v2"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func TestHandle(t *testing.T) {
	t.Parallel()

	jsonPatchType := v1beta1.PatchTypeJSONPatch

	tests := []struct {
		name             string
		mustCreateSecret bool
		target           *corev1.ServiceAccount
		patchType        *v1beta1.PatchType
		patch            []jsonpatch.JsonPatchOperation
	}{
		{
			name:             "no patch needed, must create the secret",
			mustCreateSecret: true,
			target:           newServiceAccount(1, "registry-secret"),
		},
		{
			name:             "no secrets at all, must create the secret",
			mustCreateSecret: true,
			target:           newServiceAccount(1),
			patchType:        &jsonPatchType,
			patch: []jsonpatch.JsonPatchOperation{{
				Operation: "add",
				Path:      "/imagePullSecrets",
				Value:     []interface{}{map[string]interface{}{"name": "registry-secret"}},
			}},
		},
		{
			name:             "no secrets managed by us, must create the secret",
			mustCreateSecret: true,
			target:           newServiceAccount(1, "not-managed-by-us"),
			patchType:        &jsonPatchType,
			patch: []jsonpatch.JsonPatchOperation{{
				Operation: "add",
				Path:      "/imagePullSecrets/1",
				Value:     map[string]interface{}{"name": "registry-secret"},
			}},
		},

		{
			name:             "no patch needed, must not create the secret",
			mustCreateSecret: false,
			target:           newServiceAccount(1, "registry-secret"),
		},
		{
			name:             "no secrets at all, must not create the secret",
			mustCreateSecret: false,
			target:           newServiceAccount(1),
			patchType:        &jsonPatchType,
			patch: []jsonpatch.JsonPatchOperation{{
				Operation: "add",
				Path:      "/imagePullSecrets",
				Value:     []interface{}{map[string]interface{}{"name": "registry-secret"}},
			}},
		},
		{
			name:             "no secrets managed by us, must not create the secret",
			mustCreateSecret: false,
			target:           newServiceAccount(1, "not-managed-by-us"),
			patchType:        &jsonPatchType,
			patch: []jsonpatch.JsonPatchOperation{{
				Operation: "add",
				Path:      "/imagePullSecrets/1",
				Value:     map[string]interface{}{"name": "registry-secret"},
			}},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assertMutate(t, test.target, test.patchType, test.patch, test.mustCreateSecret)
		})
	}
}

func assertMutate(t *testing.T, target *corev1.ServiceAccount, patchType *v1beta1.PatchType, patch []jsonpatch.JsonPatchOperation, mustCreateTheSecret bool) {
	t.Helper()

	secretName := types.NamespacedName{
		Namespace: "registry-secret-manager",
		Name:      "registry-secret",
	}

	var objects []client.Object
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

	// Create a client and the mutator
	fakeClient := fake.NewFakeClient(objects...)
	mutator := serviceaccount.NewMutator(fakeClient, nil)

	decoder, _ := admission.NewDecoder(scheme.Scheme)
	_ = mutator.InjectDecoder(decoder)

	// Submit the request and verify the response
	serviceAccountJSON, _ := json.Marshal(target)

	request := admission.Request{
		AdmissionRequest: v1beta1.AdmissionRequest{
			Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
			Namespace: "registry-secret-manager",
			Name:      "default",
			Object:    runtime.RawExtension{Raw: serviceAccountJSON},
		},
	}
	response := mutator.Handle(context.TODO(), request)

	assert.True(t, response.Allowed)
	assert.Equal(t, patchType, response.PatchType)
	assert.Equal(t, patch, response.Patches)

	// Check if the secret was created or updated, either way ResourceVersion should always be 1
	var desiredSecret corev1.Secret
	err := fakeClient.Get(context.TODO(), secretName, &desiredSecret)

	assert.NoError(t, err)
	assert.Equal(t, "1", desiredSecret.ResourceVersion)
}
