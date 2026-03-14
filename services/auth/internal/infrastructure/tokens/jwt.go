package tokens

import (
	"auth/internal/application/ports"
	"auth/internal/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secretKey  []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
	resetTTL   time.Duration
}

func NewJWTService(cnf config.Config) *JWTService {
	svc := &JWTService{
		secretKey:  []byte(cnf.JWTSecret),
		issuer:     cnf.JWTIssuer,
		accessTTL:  time.Duration(cnf.AccessTokenTTLMin) * time.Minute,
		refreshTTL: time.Duration(cnf.RefreshTokenTTLDays) * time.Hour * 24,
		resetTTL:   time.Duration(cnf.PasswordResetTokenTTLMin) * time.Minute,
	}
	return svc
}

func (s *JWTService) Sign(payload Payload) (string, error) {
	claims := jwt.MapClaims{
		"sub":        payload.Sub,
		"iss":        s.issuer,
		"session_id": payload.SessionID,
		"type":       payload.Type,
		"exp":        payload.Exp,
	}
	if payload.Email != "" {
		claims["email"] = payload.Email
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func parsePayload(claims jwt.MapClaims) (*Payload, error) {
	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrPayloadMalformed
	}

	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return nil, ErrPayloadMalformed
	}

	typee, ok := claims["type"].(string)
	if !ok {
		return nil, ErrPayloadMalformed
	}

	email, _ := claims["email"].(string)

	var exp int64
	switch v := claims["exp"].(type) {
	case float64:
		exp = int64(v)
	case int64:
		exp = v
	case int:
		exp = int64(v)
	default:
		return nil, ErrPayloadMalformed
	}

	return &Payload{
		Sub:       sub,
		SessionID: sessionID,
		Type:      typee,
		Exp:       exp,
		Email:     email,
	}, nil
}

func (s *JWTService) ValidateToken(tokenStr string) (*ports.TokenClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	payload, err := parsePayload(claims)
	if err != nil {
		return nil, err
	}

	if payload.Exp < time.Now().Unix() {
		return nil, ErrExpiredToken
	}

	return &ports.TokenClaims{
		Subject:   payload.Sub,
		SessionID: payload.SessionID,
		Type:      payload.Type,
		ExpiresAt: time.Unix(payload.Exp, 0),
		Email:     payload.Email,
	}, nil
}

func (s *JWTService) GenerateRefreshToken(userID, sessionID string) (string, error) {
	payload := Payload{
		Sub:       userID,
		SessionID: sessionID,
		Type:      TokenTypeRefresh,
		Exp:       time.Now().Add(s.refreshTTL).Unix(),
	}
	return s.Sign(payload)
}

func (s *JWTService) GenerateAccessToken(userID, sessionID string) (string, error) {
	payload := Payload{
		Sub:       userID,
		SessionID: sessionID,
		Type:      TokenTypeAccess,
		Exp:       time.Now().Add(s.accessTTL).Unix(),
	}

	return s.Sign(payload)
}

func (s *JWTService) GeneratePasswordResetToken(userID, email string) (string, error) {
	payload := Payload{
		Sub:       userID,
		SessionID: uuid.NewString(),
		Type:      TokenTypeReset,
		Exp:       time.Now().Add(s.resetTTL).Unix(),
		Email:     email,
	}
	return s.Sign(payload)
}

func (s *JWTService) EnsureTokenType(claims *ports.TokenClaims, tokenType string) error {
	if claims.Type != tokenType {
		return ErrInvalidTokenType
	}
	return nil
}

var _ ports.TokenManager = (*JWTService)(nil)
