package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"book-reader/backend/internal/config"
	"book-reader/backend/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	cfg *config.Config
}

type jwtClaims struct {
	Role string `json:"role"`
	Type string `json:"typ"`
	jwt.RegisteredClaims
}

func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{cfg: cfg}
}

func (s *TokenService) CreateAccessToken(user *model.User) (string, int64, error) {
	expires := time.Now().Add(s.cfg.AccessTokenTTL)
	claims := jwtClaims{
		Role: user.Role,
		Type: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.JWTSecret))
	return signed, int64(s.cfg.AccessTokenTTL.Seconds()), err
}

func (s *TokenService) VerifyAccessToken(tokenString string) (*model.AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok || claims.Type != "access" {
		return nil, fmt.Errorf("invalid claims")
	}
	return &model.AccessClaims{
		Subject: claims.Subject,
		Role:    claims.Role,
		Type:    claims.Type,
		Expires: claims.ExpiresAt.Unix(),
	}, nil
}

func NewRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, HashRefreshToken(token), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
