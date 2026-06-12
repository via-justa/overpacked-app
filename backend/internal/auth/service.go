package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrInvalidToken = errors.New("invalid token")

type Service struct {
	username string
	password string
	secret   []byte
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

type claims struct {
	Username string `json:"username"`
	Type     string `json:"typ"`
	jwt.RegisteredClaims
}

func NewService(username, password, jwtSecret string) (*Service, error) {
	if username == "" {
		return nil, fmt.Errorf("auth username is required")
	}
	if password == "" {
		return nil, fmt.Errorf("auth password is required")
	}
	if jwtSecret == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}

	return &Service{
		username: username,
		password: password,
		secret:   []byte(jwtSecret),
	}, nil
}

func (s *Service) Login(username, password string) (*TokenPair, error) {
	if subtle.ConstantTimeCompare([]byte(username), []byte(s.username)) != 1 ||
		subtle.ConstantTimeCompare([]byte(password), []byte(s.password)) != 1 {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(username)
}

func (s *Service) Refresh(refreshToken string) (*TokenPair, error) {
	parsedClaims, err := s.parseAndValidate(refreshToken, "refresh")
	if err != nil {
		return nil, err
	}

	return s.issueTokens(parsedClaims.Username)
}

func (s *Service) Logout(_ string) error {
	// Stateless JWT logout; token invalidation can be added later via denylist.
	return nil
}

// ValidateAccessToken returns ErrInvalidToken unless raw is a valid, unexpired access token.
func (s *Service) ValidateAccessToken(raw string) error {
	_, err := s.parseAndValidate(raw, "access")
	return err
}

func (s *Service) issueTokens(username string) (*TokenPair, error) {
	accessToken, err := s.signToken(username, "access", accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.signToken(username, "refresh", refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessTokenTTL.Seconds()),
	}, nil
}

func (s *Service) signToken(username, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := claims{
		Username: username,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func (s *Service) parseAndValidate(rawToken, expectedType string) (*claims, error) {
	parsed := &claims{}
	token, err := jwt.ParseWithClaims(rawToken, parsed, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	if parsed.Type != expectedType {
		return nil, ErrInvalidToken
	}

	return parsed, nil
}
