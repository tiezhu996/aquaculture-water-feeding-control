package service

import (
	"aquaculture-water-feeding-control/backend/internal/dto"
	"aquaculture-water-feeding-control/backend/internal/repository"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TokenClaims struct {
	UserID      uint   `json:"uid"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

type AuthService struct {
	users  *repository.UserRepository
	secret []byte
	ttl    time.Duration
}

func NewAuthService(users *repository.UserRepository, secret string, ttl time.Duration) *AuthService {
	return &AuthService{users: users, secret: []byte(secret), ttl: ttl}
}

func (s *AuthService) Login(input dto.LoginRequest) (dto.LoginResponse, error) {
	user, err := s.users.FindByUsername(input.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.LoginResponse{}, NewError(CodeUnauthorized, "用户名或密码错误")
		}
		return dto.LoginResponse{}, WrapError(CodeInternal, "查询用户失败", err)
	}
	if !user.Active || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return dto.LoginResponse{}, NewError(CodeUnauthorized, "用户名或密码错误")
	}
	now := time.Now().UTC()
	claims := TokenClaims{
		UserID:      user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "aquaculture-control",
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return dto.LoginResponse{}, WrapError(CodeInternal, "签发访问令牌失败", err)
	}
	return dto.LoginResponse{
		Token: signed,
		User:  dto.UserResponse{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Role: string(user.Role)},
	}, nil
}

func (s *AuthService) ParseToken(raw string) (TokenClaims, error) {
	claims := TokenClaims{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method %s", token.Method.Alg())
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return TokenClaims{}, NewError(CodeUnauthorized, "访问令牌无效或已过期")
	}
	user, err := s.users.FindByID(claims.UserID)
	if err != nil || !user.Active {
		return TokenClaims{}, NewError(CodeUnauthorized, "用户不存在或已停用")
	}
	if string(user.Role) != claims.Role {
		return TokenClaims{}, NewError(CodeUnauthorized, "用户角色已变更，请重新登录")
	}
	return claims, nil
}

func (s *AuthService) Me(userID uint) (dto.UserResponse, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.UserResponse{}, NewError(CodeNotFound, "用户不存在")
		}
		return dto.UserResponse{}, WrapError(CodeInternal, "查询用户失败", err)
	}
	return dto.UserResponse{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName, Role: string(user.Role)}, nil
}
