package service_test

import (
	"testing"
	"time"

	"aquaculture-water-feeding-control/backend/internal/database"
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"aquaculture-water-feeding-control/backend/internal/service"
)

func newAuthService(t *testing.T) *service.AuthService {
	t.Helper()
	db, err := database.Open("file:authscoring?mode=memory&cache=shared", "sqlite", "production")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return service.NewAuthService(repository.NewUserRepository(db), "auth-scoring-secret-2026-change-me", 8*time.Hour)
}

func assertCode(t *testing.T, err error, want service.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %s, got nil", want)
	}
	appErr, ok := service.AsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if appErr.Code != want {
		t.Fatalf("error code = %s, want %s", appErr.Code, want)
	}
}

func TestLoginUnknownUserReturnsUnauthorized(t *testing.T) {
	svc := newAuthService(t)
	_, err := svc.Login(dto.LoginRequest{Username: "ghostuser", Password: "whatever123"})
	assertCode(t, err, service.CodeUnauthorized)
}

func TestMeMissingUserReturnsNotFound(t *testing.T) {
	svc := newAuthService(t)
	_, err := svc.Me(999999)
	assertCode(t, err, service.CodeNotFound)
}
