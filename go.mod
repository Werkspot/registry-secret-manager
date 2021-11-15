module registry-secret-manager

go 1.15

require (
	github.com/aws/aws-sdk-go v1.37.26
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	gomodules.xyz/jsonpatch/v2 v2.2.0
	k8s.io/api v0.23.0-alpha.3
	k8s.io/apimachinery v0.23.0-alpha.3
	k8s.io/client-go v0.23.0-alpha.3
	sigs.k8s.io/controller-runtime v0.11.0-beta.0
)
