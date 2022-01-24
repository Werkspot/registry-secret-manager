module registry-secret-manager

go 1.15

require (
	github.com/aws/aws-sdk-go v1.37.26
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	gomodules.xyz/jsonpatch/v2 v2.1.0
	k8s.io/api v0.23.2
	k8s.io/apimachinery v0.23.2
	k8s.io/client-go v0.23.2
	sigs.k8s.io/controller-runtime v0.7.0-alpha.6
)
