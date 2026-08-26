package config

import "testing"

func TestProductionRejectsDefaultJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "local-development-secret-change-me")
	if _, err := Load(); err == nil {
		t.Fatal("production must reject the default JWT secret")
	}
}

func TestDevelopmentAllowsDocumentedLocalSecret(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("JWT_SECRET", "local-development-secret-change-me")
	if _, err := Load(); err != nil {
		t.Fatalf("development config rejected: %v", err)
	}
}
