package serviceaccount

import (
	corev1 "k8s.io/api/core/v1"
)

// Check if the ServiceAccount needs mutation
func needsMutation(serviceAccount *corev1.ServiceAccount) bool {
	for _, imagePullSecret := range serviceAccount.ImagePullSecrets {
		if imagePullSecret.Name == "registry-secret" {
			return false
		}
	}

	return true
}
