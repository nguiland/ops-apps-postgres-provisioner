package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestRequiredCredentials_DatabaseNameIsMissing_DefaultToUsername(t *testing.T) {
	secret := &corev1.Secret{Data: map[string][]byte{
		usernameKey: []byte("app_user"),
		passwordKey: []byte("secret-password"),
	}}

	database, username, password, err := requiredCredentials(secret, true)
	if err != nil {
		t.Fatalf("requiredCredentials returned error: %v", err)
	}

	if database != "app_user" || username != "app_user" || password != "secret-password" {
		t.Fatalf(
			"unexpected credentials: database=%q username=%q password length=%d",
			database,
			username,
			len(password),
		)
	}
}

func TestRequiredCredentials_DatabaseNameIsProvided_UseExplicitDatabaseName(t *testing.T) {
	secret := &corev1.Secret{Data: map[string][]byte{
		databaseNameKey: []byte("app_database"),
		usernameKey:     []byte("app_user"),
		passwordKey:     []byte("secret-password"),
	}}

	database, username, _, err := requiredCredentials(secret, true)
	if err != nil {
		t.Fatalf("requiredCredentials returned error: %v", err)
	}

	if database != "app_database" || username != "app_user" {
		t.Fatalf("unexpected database/user: %q/%q", database, username)
	}
}

func TestRequiredCredentials_PasswordIsMissing_ReturnError(t *testing.T) {
	secret := &corev1.Secret{Data: map[string][]byte{
		usernameKey: []byte("app_user"),
	}}

	if _, _, _, err := requiredCredentials(secret, true); err == nil {
		t.Fatal("expected missing password to fail")
	}
}
