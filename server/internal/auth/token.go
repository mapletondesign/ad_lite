package auth

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenDuration = 15 * time.Minute
	DeviceTokenDuration = 365 * 24 * time.Hour
)

type Claims struct {
	jwt.RegisteredClaims
	Role         string `json:"role"`
	AdvertiserID string `json:"advertiser_id,omitempty"`
	VenueID      string `json:"venue_id,omitempty"`
}

func IssueDeviceToken(privateKey *rsa.PrivateKey, deviceID string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   deviceID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(DeviceTokenDuration)),
		},
		Role: "device",
	})
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("sign device token: %w", err)
	}
	return signed, nil
}

func IssueAccessToken(privateKey *rsa.PrivateKey, userID, role, advertiserID, venueID string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenDuration)),
		},
		Role:         role,
		AdvertiserID: advertiserID,
		VenueID:      venueID,
	})
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func VerifyToken(publicKey *rsa.PublicKey, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}
	return claims, nil
}

// VerifyTokenAllowExpired verifies the RS256 signature but does not reject expired tokens.
// Used only for device re-authentication after a long-lived token expires.
func VerifyTokenAllowExpired(publicKey *rsa.PublicKey, tokenStr string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}
	return claims, nil
}
