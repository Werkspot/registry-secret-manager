package main

import (
	"os"
	"registry-secret-manager/cmd"
)

func main() {
	os.Exit(cmd.NewRegistrySecretManager().Run())
}
