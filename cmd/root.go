package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"registry-secret-manager/pkg/registry"
	"registry-secret-manager/pkg/secret"
	"registry-secret-manager/pkg/serviceaccount"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Config holds the application configuration
type Config struct{}

// RegistrySecretManager main application
type RegistrySecretManager struct {
	config  Config
	command *cobra.Command
}

// NewRegistrySecretManager returns a pointer to RegistrySecretManager
func NewRegistrySecretManager() *RegistrySecretManager {
	cfg, err := readConfig()
	if err != nil {
		panic(err)
	}

	return &RegistrySecretManager{
		config:  cfg,
		command: getCommand(),
	}
}

// Run the main application
func (app *RegistrySecretManager) Run() int {
	app.command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return app.initLogger()
	}

	if err := app.command.Execute(); err != nil {
		log.Error(err)
		return 1
	}

	return 0
}

func (app *RegistrySecretManager) initLogger() (err error) {
	level, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		return
	}

	log.SetOutput(os.Stdout)
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation: true,
		ForceColors:            true,
	})

	return
}

func readConfig() (config Config, err error) {
	ex, err := os.Executable()
	if err != nil {
		return
	}

	home, err := homedir.Dir()
	if err != nil {
		return
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Dir(ex))
	viper.AddConfigPath(home)
	viper.SetConfigName("config")
	viper.SetConfigType("yml")

	viper.SetEnvPrefix("REGISTRY_SECRET_MANAGER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		return
	}
	if err = viper.Unmarshal(&config); err != nil {
		return
	}

	return
}

func bindFlags(flag *pflag.Flag) {
	viper.RegisterAlias(strings.ReplaceAll(flag.Name, "-", "_"), flag.Name)
}

func getCommand() (c *cobra.Command) {
	pflag.String("cert-dir", "", "Directory that holds the tls.crt and tls.key files")
	pflag.String("log-level", "warning", "Log verbosity level")
	pflag.Parse()

	if err := viper.BindPFlags(pflag.CommandLine); err != nil {
		panic(err)
	}

	pflag.VisitAll(bindFlags)

	return &cobra.Command{
		Use:   "registry-secret-manager",
		Short: "Manages the creation and distribution of credentials for container registries",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// Setup the container registries
			registries := []registry.Registry{
				registry.NewDockerHub(),
				registry.NewECR(),
			}

			// Setup the manager
			mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
				Host:    "",
				Port:    8443,
				CertDir: viper.GetString("cert-dir"),

				HealthProbeBindAddress: ":8080",
				MetricsBindAddress:     ":8081",

				LeaderElection:             true,
				LeaderElectionID:           "registry-secret-manager",
				LeaderElectionNamespace:    "registry-secret-manager",
				LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
			})
			if err != nil {
				err = fmt.Errorf("unable to set up overall controller manager: %v", err)
				return
			}

			// Add healthz and readyz check
			err = mgr.AddHealthzCheck("ping", healthz.Ping)
			if err != nil {
				err = fmt.Errorf("failed to add ping healthz check")
				return
			}

			err = mgr.AddReadyzCheck("ping", healthz.Ping)
			if err != nil {
				err = fmt.Errorf("failed to add ping readyz check")
				return
			}

			// Setup a new controller to reconcile ServiceAccounts
			err = serviceaccount.NewController(mgr, registries)
			if err != nil {
				return
			}

			// Setup a new controller to reconcile Secrets
			err = secret.NewController(mgr, registries)
			if err != nil {
				return
			}

			// Start the controller manager
			log.Infof("Starting controller manager")

			err = mgr.Start(signals.SetupSignalHandler())
			if err != nil {
				err = fmt.Errorf("unable to start manager: %v", err)
				return
			}

			return
		},
	}
}
