package internal

import (
	"os"

	"github.com/MarioCarrion/todo-api-microservice-example/internal"
	"github.com/MarioCarrion/todo-api-microservice-example/internal/envvar/vault"
)

// NewVaultProvider instantiates the Vault client using configuration defined in environment variables.
func NewVaultProvider() (*vault.Provider, error) {
	// XXX: We will revisit this code in future episodes replacing it with another solution
	vaultPath := os.Getenv("VAULT_PATH")
	vaultToken := os.Getenv("VAULT_TOKEN")
	vaultAddress := os.Getenv("VAULT_ADDRESS")
	// XXX: -

	provider, err := vault.New(vaultToken, vaultAddress, vaultPath)
	if err != nil {
		return nil, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "vault.New ")
	}

	return provider, nil
}
