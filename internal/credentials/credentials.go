package credentials

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const service = "pb-agent"

func Save(reference, token string) error {
	if reference == "" || token == "" {
		return fmt.Errorf("credential reference and token are required")
	}
	return keyring.Set(service, reference, token)
}

func Delete(reference string) error {
	if reference == "" {
		return nil
	}
	err := keyring.Delete(service, reference)
	if err == keyring.ErrNotFound {
		return nil
	}
	return err
}

func Resolve(reference string) (string, error) {
	if token := os.Getenv("PB_AGENT_TOKEN"); token != "" {
		return token, nil
	}
	if reference == "" {
		return "", nil
	}
	token, err := keyring.Get(service, reference)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("credential %q was not found in the OS keychain", reference)
	}
	return token, err
}
