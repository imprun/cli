package impruncli

import (
	"errors"
	"testing"

	"github.com/zalando/go-keyring"
)

func TestCredentialStoreMigratesLegacyWFEntry(t *testing.T) {
	originalGet := getCredential
	originalSet := setCredential
	originalDelete := deleteCredential
	t.Cleanup(func() {
		getCredential = originalGet
		setCredential = originalSet
		deleteCredential = originalDelete
	})

	values := map[string]string{
		legacyKeyringService + "/cell.example/user@example.com": "legacy-token",
	}
	getCredential = func(service, key string) (string, error) {
		value, ok := values[service+"/"+key]
		if !ok {
			return "", keyring.ErrNotFound
		}
		return value, nil
	}
	setCredential = func(service, key, value string) error {
		values[service+"/"+key] = value
		return nil
	}
	deleteCredential = func(service, key string) error {
		delete(values, service+"/"+key)
		return nil
	}

	store := credentialStore{}
	value, found, err := store.Get("cell.example/user@example.com")
	if err != nil || !found || value != "legacy-token" {
		t.Fatalf("Get() value=%q found=%v err=%v", value, found, err)
	}
	if values[keyringService+"/cell.example/user@example.com"] != "legacy-token" {
		t.Fatal("legacy credential was not migrated to the imprun service")
	}

	if err := store.Delete("cell.example/user@example.com"); err != nil {
		t.Fatalf("Delete() error=%v", err)
	}
	for _, service := range []string{keyringService, legacyKeyringService} {
		if _, ok := values[service+"/cell.example/user@example.com"]; ok {
			t.Fatalf("credential remains in %s service", service)
		}
	}
}

func TestCredentialStoreDoesNotHideMigrationFailure(t *testing.T) {
	originalGet := getCredential
	originalSet := setCredential
	t.Cleanup(func() {
		getCredential = originalGet
		setCredential = originalSet
	})

	getCredential = func(service, key string) (string, error) {
		if service == legacyKeyringService {
			return "legacy-token", nil
		}
		return "", keyring.ErrNotFound
	}
	setCredential = func(service, key, value string) error {
		return errors.New("write failed")
	}

	if _, found, err := (credentialStore{}).Get("cell.example/user@example.com"); err == nil || found {
		t.Fatalf("Get() found=%v err=%v", found, err)
	}
}
