package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mapletondesign/ad_lite/api"
	"github.com/mapletondesign/ad_lite/internal/auth"
	"github.com/mapletondesign/ad_lite/internal/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	pool := testhelper.DB(t)
	rdb := testhelper.Redis(t)
	priv, pub := testhelper.RSAKeys(t)
	testhelper.TruncateAll(t, pool)
	return api.NewRouter(pool, rdb, priv, pub, nil)
}

func postJSON(t *testing.T, handler http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func getWithToken(t *testing.T, handler http.Handler, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func TestHealth(t *testing.T) {
	h := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, true, body["db"])
	assert.Equal(t, true, body["redis"])
}

func TestAuthRegisterAndLogin(t *testing.T) {
	h := newTestServer(t)

	t.Run("register returns 201 with token pair", func(t *testing.T) {
		w := postJSON(t, h, "/api/v1/auth/register", map[string]string{
			"email":    "newuser@example.com",
			"password": "password123",
			"role":     "advertiser",
			"name":     "New User",
		})
		assert.Equal(t, http.StatusCreated, w.Code)
		var pair map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&pair))
		assert.NotEmpty(t, pair["access_token"])
		assert.NotEmpty(t, pair["refresh_token"])
	})

	t.Run("login with valid credentials returns 200", func(t *testing.T) {
		w := postJSON(t, h, "/api/v1/auth/login", map[string]string{
			"email":    "newuser@example.com",
			"password": "password123",
		})
		assert.Equal(t, http.StatusOK, w.Code)
		var pair map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&pair))
		assert.NotEmpty(t, pair["access_token"])
	})

	t.Run("login with wrong password returns 401", func(t *testing.T) {
		w := postJSON(t, h, "/api/v1/auth/login", map[string]string{
			"email":    "newuser@example.com",
			"password": "wrongpassword",
		})
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("register with invalid role returns 400", func(t *testing.T) {
		w := postJSON(t, h, "/api/v1/auth/register", map[string]string{
			"email":    "admin@example.com",
			"password": "password123",
			"role":     "admin",
			"name":     "Admin",
		})
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestVenueRouteRoleEnforcement(t *testing.T) {
	pool := testhelper.DB(t)
	rdb := testhelper.Redis(t)
	priv, pub := testhelper.RSAKeys(t)
	testhelper.TruncateAll(t, pool)
	h := api.NewRouter(pool, rdb, priv, pub, nil)

	// Mint tokens directly for each role so we're not dependent on the register flow.
	adminToken, err := auth.IssueAccessToken(priv, "user-admin", "admin", "", "")
	require.NoError(t, err)

	advToken, err := auth.IssueAccessToken(priv, "user-adv", "advertiser", "adv-123", "")
	require.NoError(t, err)

	t.Run("no token returns 401", func(t *testing.T) {
		w := getWithToken(t, h, "/api/v1/venues", "")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("advertiser token returns 403", func(t *testing.T) {
		w := getWithToken(t, h, "/api/v1/venues", advToken)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("admin token returns 200", func(t *testing.T) {
		w := getWithToken(t, h, "/api/v1/venues", adminToken)
		assert.Equal(t, http.StatusOK, w.Code)
	})

}

func TestDeviceRegisterHTTP(t *testing.T) {
	pool := testhelper.DB(t)
	rdb := testhelper.Redis(t)
	priv, pub := testhelper.RSAKeys(t)
	testhelper.TruncateAll(t, pool)
	ctx := context.Background()
	h := api.NewRouter(pool, rdb, priv, pub, nil)

	var venueID string
	require.NoError(t, pool.QueryRow(ctx,
		`INSERT INTO venues (name) VALUES ('HTTP Venue') RETURNING id`,
	).Scan(&venueID))

	t.Run("device register returns 201 with device_id and token", func(t *testing.T) {
		w := postJSON(t, h, "/api/v1/devices/register", map[string]string{
			"venue_id":         venueID,
			"name":             "HTTP Screen",
			"firmware_version": "1.0.0",
		})
		assert.Equal(t, http.StatusCreated, w.Code)
		var resp map[string]any
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.NotEmpty(t, resp["device_id"])
		assert.NotEmpty(t, resp["token"])
	})
}

func TestHeartbeatRequiresDeviceJWT(t *testing.T) {
	pool := testhelper.DB(t)
	rdb := testhelper.Redis(t)
	priv, pub := testhelper.RSAKeys(t)
	testhelper.TruncateAll(t, pool)
	h := api.NewRouter(pool, rdb, priv, pub, nil)

	// Mint a user (admin) token — must be rejected on a device route.
	userToken, err := auth.IssueAccessToken(priv, "user-123", "admin", "", "")
	require.NoError(t, err)

	t.Run("no token returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/fake-id/heartbeat", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("user JWT on device route returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/fake-id/heartbeat", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
