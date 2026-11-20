package controller

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func requiredCredentials(
	secret *corev1.Secret,
	includesDatabase bool,
) (database, username, password string, err error) {
	username = strings.TrimSpace(string(secret.Data[usernameKey]))
	password = string(secret.Data[passwordKey])

	if username == "" {
		return "", "", "", fmt.Errorf("key %q is missing or blank", usernameKey)
	}

	if len(username) > 63 || strings.ContainsRune(username, '\x00') {
		return "", "", "", fmt.Errorf(
			"key %q must be a valid PostgreSQL identifier of at most 63 bytes",
			usernameKey,
		)
	}

	if strings.TrimSpace(password) == "" || strings.ContainsRune(password, '\x00') {
		return "", "", "", fmt.Errorf(
			"key %q is missing, blank, or contains a NUL byte",
			passwordKey,
		)
	}

	if includesDatabase {
		database = strings.TrimSpace(string(secret.Data[databaseNameKey]))

		if database == "" {
			database = username
		}

		if len(database) > 63 || strings.ContainsRune(database, '\x00') {
			return "", "", "", fmt.Errorf(
				"key %q must be a valid PostgreSQL identifier of at most 63 bytes",
				databaseNameKey,
			)
		}
	}

	return database, username, password, nil
}
