package serviceaccount

import (
	"context"
	"encoding/json"
	"testing"

	"registry-secret-manager/pkg/secret"

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
	jsonPatchType := v1beta1.PatchTypeJSONPatch

	tests := map[string]struct {
		target    *corev1.ServiceAccount
		patchType *v1beta1.PatchType
		patch     []jsonpatch.JsonPatchOperation
	}{
		"no patch needed": {
			target: newServiceAccount(1, "registry-secret"),
		},
		"no secrets at all": {
			target:    newServiceAccount(1),
			patchType: &jsonPatchType,
			patch: []jsonpatch.JsonPatchOperation{{
				Operation: "add",
				Path:      "/imagePullSecrets",
				Value:     []interface{}{map[string]interface{}{"name": "registry-secret"}},
			}},
		},
		"no secrets managed by us": {
			target:    newServiceAccount(1, "not-managed-by-us"),
			patchType: &jsonPatchType,
			patch: []jsonpatch.JsonPatchOperation{{
				Operation: "add",
				Path:      "/imagePullSecrets/1",
				Value:     map[string]interface{}{"name": "registry-secret"},
			}},
		},
	}

	for name, test := range tests {
		t.Run(name+", must create the secret", func(t *testing.T) {
			assertMutate(t, test.target, test.patchType, test.patch, true)
		})

		t.Run(name+", must not create the secret", func(t *testing.T) {
			assertMutate(t, test.target, test.patchType, test.patch, false)
		})
	}
}

func assertMutate(t *testing.T, target *corev1.ServiceAccount, patchType *v1beta1.PatchType, patch []jsonpatch.JsonPatchOperation, mustCreateTheSecret bool) {
	secretName := types.NamespacedName{
		Namespace: "registry-secret-manager",
		Name:      "registry-secret",
	}

	var objects []client.Object
	if !mustCreateTheSecret {
		// Secret is not created by the mutator
		objects = append(objects, &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: secret.SecretResource.Version,
				Kind:       secret.SecretResource.Kind,
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
	mutator := newMutator(fakeClient, nil)

	decoder, _ := admission.NewDecoder(scheme.Scheme)
	_ = mutator.InjectDecoder(decoder)

	// Submit the request and verify the response
	serviceAccountJson, _ := json.Marshal(target)

	request := admission.Request{
		AdmissionRequest: v1beta1.AdmissionRequest{
			Kind:      metav1.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
			Namespace: "registry-secret-manager",
			Name:      "default",
			Object:    runtime.RawExtension{Raw: serviceAccountJson},
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
