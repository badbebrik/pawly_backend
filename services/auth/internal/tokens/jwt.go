package tokens

import (
	"auth/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTService struct {
	secretKey  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTService(cnf config.Config) *JWTService {
	svc := &JWTService{
		secretKey:  []byte(cnf.JWTSecret),
		accessTTL:  time.Duration(cnf.AccessTokenTTLMin) * time.Minute,
		refreshTTL: time.Duration(cnf.RefreshTokenTTLDays) * time.Hour * 24,
	}
	return svc
}

func (s *JWTService) Sign(payload Payload) (string, error) {
	claims := jwt.MapClaims{
		"sub":        payload.Sub,
		"session_id": payload.SessionID,
		"type":       payload.Type,
		"exp":        payload.Exp,
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

	exp, ok := claims["exp"].(int64)
	if !ok {
		return nil, ErrPayloadMalformed
	}

	return &Payload{
		Sub:       sub,
		SessionID: sessionID,
		Type:      typee,
		Exp:       exp,
	}, nil
}

func (s *JWTService) ValidateToken(tokenStr string) (*Payload, error) {
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

	return payload, nil
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
		SessionID: email,
		Type:      TokenTypeReset,
		Exp:       time.Now().Add(time.Minute * 15).Unix(), // TODO: Просунуть в конфиг
	}
	return s.Sign(payload)
}

func (s *JWTService) EnsureTokenType(p *Payload, tType string) error {
	if p.Type != tType {
		return ErrInvalidTokenType
	}
	return nil
}
