package impruncli

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const (
	keyringService       = "imprun"
	legacyKeyringService = "wf"
)

var (
	getCredential    = keyring.Get
	setCredential    = keyring.Set
	deleteCredential = keyring.Delete
)

type credentialStore struct{}

func (credentialStore) Get(key string) (string, bool, error) {
	value, err := getCredential(keyringService, key)
	if err == nil {
		return value, true, nil
	}
	if !errors.Is(err, keyring.ErrNotFound) {
		return "", false, err
	}

	value, err = getCredential(legacyKeyringService, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if err := setCredential(keyringService, key, value); err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (credentialStore) Set(key, value string) error {
	return setCredential(keyringService, key, value)
}

func (credentialStore) Delete(key string) error {
	for _, service := range []string{keyringService, legacyKeyringService} {
		err := deleteCredential(service, key)
		if err != nil && !errors.Is(err, keyring.ErrNotFound) {
			return err
		}
	}
	return nil
}
