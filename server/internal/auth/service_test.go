package auth_test

import (
	"context"
	"testing"

	"github.com/mapletondesign/ad_lite/internal/auth"
	"github.com/mapletondesign/ad_lite/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	pool := testhelper.DB(t)
	priv, pub := testhelper.RSAKeys(t)

	t.Run("advertiser role creates advertiser row and user row", func(t *testing.T) {
		testhelper.TruncateAll(t, pool)
		svc := auth.NewService(pool, priv, pub)

		pair, err := svc.Register(context.Background(), auth.RegisterRequest{
			Email:    "advertiser@example.com",
			Password: "supersecret",
			Role:     "advertiser",
			Name:     "Acme Corp",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
		assert.Greater(t, pair.ExpiresIn, 0)

		// Verify an advertiser row was created.
		var count int
		require.NoError(t, pool.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM advertisers WHERE email = 'advertiser@example.com'`,
		).Scan(&count))
		assert.Equal(t, 1, count)

		// Verify a user row with the correct role was created.
		var role string
		require.NoError(t, pool.QueryRow(context.Background(),
			`SELECT role FROM users WHERE email = 'advertiser@example.com'`,
		).Scan(&role))
		assert.Equal(t, "advertiser", role)
	})

	t.Run("venue role creates venue row and user row", func(t *testing.T) {
		testhelper.TruncateAll(t, pool)
		svc := auth.NewService(pool, priv, pub)

		pair, err := svc.Register(context.Background(), auth.RegisterRequest{
			Email:    "venue@example.com",
			Password: "supersecret",
			Role:     "venue",
			Name:     "The Coffee Shop",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, pair.AccessToken)

		var count int
		require.NoError(t, pool.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM venues WHERE name = 'The Coffee Shop'`,
		).Scan(&count))
		assert.Equal(t, 1, count)
	})

	t.Run("invalid role is rejected", func(t *testing.T) {
		testhelper.TruncateAll(t, pool)
		svc := auth.NewService(pool, priv, pub)

		_, err := svc.Register(context.Background(), auth.RegisterRequest{
			Email:    "admin@example.com",
			Password: "secret",
			Role:     "admin",
			Name:     "Superuser",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "role must be")
	})

	t.Run("missing required fields", func(t *testing.T) {
		testhelper.TruncateAll(t, pool)
		svc := auth.NewService(pool, priv, pub)

		_, err := svc.Register(context.Background(), auth.RegisterRequest{
			Email: "nopw@example.com",
			Role:  "advertiser",
			Name:  "No Password",
		})
		require.Error(t, err)
	})

	t.Run("duplicate email is rejected", func(t *testing.T) {
		testhelper.TruncateAll(t, pool)
		svc := auth.NewService(pool, priv, pub)
		ctx := context.Background()

		_, err := svc.Register(ctx, auth.RegisterRequest{
			Email:    "dup@example.com",
			Password: "secret",
			Role:     "advertiser",
			Name:     "First",
		})
		require.NoError(t, err)

		_, err = svc.Register(ctx, auth.RegisterRequest{
			Email:    "dup@example.com",
			Password: "secret2",
			Role:     "advertiser",
			Name:     "Second",
		})
		require.Error(t, err)
	})
}

func TestLogin(t *testing.T) {
	pool := testhelper.DB(t)
	priv, pub := testhelper.RSAKeys(t)
	testhelper.TruncateAll(t, pool)
	svc := auth.NewService(pool, priv, pub)

	_, err := svc.Register(context.Background(), auth.RegisterRequest{
		Email:    "login@example.com",
		Password: "correct-password",
		Role:     "advertiser",
		Name:     "Login Test",
	})
	require.NoError(t, err)

	t.Run("valid credentials return token pair", func(t *testing.T) {
		pair, err := svc.Login(context.Background(), auth.LoginRequest{
			Email:    "login@example.com",
			Password: "correct-password",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)

		// The access token must decode to the correct role.
		claims, err := auth.VerifyToken(pub, pair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, "advertiser", claims.Role)
	})

	t.Run("wrong password returns error", func(t *testing.T) {
		_, err := svc.Login(context.Background(), auth.LoginRequest{
			Email:    "login@example.com",
			Password: "wrong-password",
		})
		require.Error(t, err)
		// Error message must not reveal whether the email exists.
		assert.Equal(t, "invalid credentials", err.Error())
	})

	t.Run("unknown email returns same error as wrong password", func(t *testing.T) {
		_, err := svc.Login(context.Background(), auth.LoginRequest{
			Email:    "nobody@example.com",
			Password: "anything",
		})
		require.Error(t, err)
		assert.Equal(t, "invalid credentials", err.Error())
	})
}

func TestRefresh(t *testing.T) {
	pool := testhelper.DB(t)
	priv, pub := testhelper.RSAKeys(t)
	testhelper.TruncateAll(t, pool)
	svc := auth.NewService(pool, priv, pub)
	ctx := context.Background()

	first, err := svc.Register(ctx, auth.RegisterRequest{
		Email:    "refresh@example.com",
		Password: "secret",
		Role:     "advertiser",
		Name:     "Refresh Test",
	})
	require.NoError(t, err)

	t.Run("valid refresh token returns new token pair", func(t *testing.T) {
		second, err := svc.Refresh(ctx, auth.RefreshRequest{RefreshToken: first.RefreshToken})
		require.NoError(t, err)
		assert.NotEmpty(t, second.AccessToken)
		assert.NotEmpty(t, second.RefreshToken)
		// Refresh token must be rotated (uses crypto/rand, always unique).
		// Access token may be equal if both were issued within the same JWT-timestamp second.
		assert.NotEqual(t, first.RefreshToken, second.RefreshToken)
	})

	t.Run("reusing a consumed refresh token is rejected", func(t *testing.T) {
		// The original token was consumed by the previous sub-test.
		_, err := svc.Refresh(ctx, auth.RefreshRequest{RefreshToken: first.RefreshToken})
		require.Error(t, err)
	})

	t.Run("random token string is rejected", func(t *testing.T) {
		_, err := svc.Refresh(ctx, auth.RefreshRequest{RefreshToken: "not-a-real-token"})
		require.Error(t, err)
	})
}
