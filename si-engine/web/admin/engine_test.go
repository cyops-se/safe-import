package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	db "github.com/cyops-se/safe-import/si-engine/web/admin/db"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	if err := db.ConnectDatabase(filepath.Join(t.TempDir(), "test.db")); err != nil {
		t.Fatalf("ConnectDatabase: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB.DB(); err == nil {
			sqlDB.Close()
		}
	})
	cfg := &Config{
		ListenAddr:   ":0",
		JWTKey:       []byte("0123456789abcdef0123456789abcdef-test"),
		TokenTimeout: time.Hour,
	}
	r, err := NewRouter(cfg, nil)
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	return r
}

func doJSON(t *testing.T, r http.Handler, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func loginAs(t *testing.T, r http.Handler, email, password string) string {
	t.Helper()
	w := doJSON(t, r, http.MethodPost, "/api/auth/login", "", gin.H{"email": email, "password": password})
	if w.Code != http.StatusOK {
		t.Fatalf("login %s: expected 200, got %d: %s", email, w.Code, w.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Token == "" {
		t.Fatalf("login %s: no token in %s", email, w.Body.String())
	}
	return resp.Token
}

func TestRegistrationOnlyForFirstUser(t *testing.T) {
	r := newTestRouter(t)

	first := gin.H{"email": "admin@example.com", "password": "correct horse battery", "fullname": "Admin"}
	if w := doJSON(t, r, http.MethodPost, "/api/auth/register", "", first); w.Code != http.StatusOK {
		t.Fatalf("first registration: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var user db.User
	db.DB.Where("user_name = ?", "admin@example.com").First(&user)
	if user.Role != db.RoleAdmin {
		t.Errorf("first user should be admin, got %q", user.Role)
	}
	if !db.IsPasswordHash(user.Password) {
		t.Errorf("password must be stored as a bcrypt hash, got %q", user.Password)
	}

	second := gin.H{"email": "evil@example.com", "password": "correct horse battery", "fullname": "Evil"}
	if w := doJSON(t, r, http.MethodPost, "/api/auth/register", "", second); w.Code != http.StatusForbidden {
		t.Fatalf("second registration: expected 403, got %d", w.Code)
	}
}

func TestRegistrationRejectsShortPassword(t *testing.T) {
	r := newTestRouter(t)
	w := doJSON(t, r, http.MethodPost, "/api/auth/register", "", gin.H{"email": "a@example.com", "password": "short", "fullname": "A"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestLoginAndRoles(t *testing.T) {
	r := newTestRouter(t)
	doJSON(t, r, http.MethodPost, "/api/auth/register", "", gin.H{"email": "admin@example.com", "password": "correct horse battery", "fullname": "Admin"})

	// Wrong password
	if w := doJSON(t, r, http.MethodPost, "/api/auth/login", "", gin.H{"email": "admin@example.com", "password": "wrong password!"}); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password: expected 401, got %d", w.Code)
	}

	adminToken := loginAs(t, r, "admin@example.com", "correct horse battery")

	// Admin creates a viewer
	w := doJSON(t, r, http.MethodPost, "/api/user", adminToken, gin.H{"email": "viewer@example.com", "password": "viewer password", "fullname": "Viewer", "role": "viewer"})
	if w.Code != http.StatusOK {
		t.Fatalf("create viewer: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "$2a$") {
		t.Errorf("password hash must not be returned: %s", w.Body.String())
	}

	viewerToken := loginAs(t, r, "viewer@example.com", "viewer password")

	// Viewer may read
	if w := doJSON(t, r, http.MethodGet, "/api/user", viewerToken, nil); w.Code != http.StatusOK {
		t.Errorf("viewer GET: expected 200, got %d", w.Code)
	}

	// ... but not write
	if w := doJSON(t, r, http.MethodPost, "/api/user", viewerToken, gin.H{"email": "x@example.com", "password": "some password", "fullname": "X"}); w.Code != http.StatusForbidden {
		t.Errorf("viewer POST: expected 403, got %d", w.Code)
	}

	// ... except changing the own password
	if w := doJSON(t, r, http.MethodPut, "/api/password", viewerToken, gin.H{"password": "viewer password", "newpassword": "new viewer password"}); w.Code != http.StatusOK {
		t.Errorf("viewer change own password: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	loginAs(t, r, "viewer@example.com", "new viewer password")

	// Without token
	if w := doJSON(t, r, http.MethodGet, "/api/user", "", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("no token: expected 401, got %d", w.Code)
	}
}

func TestLastAdminCanNotBeRemoved(t *testing.T) {
	r := newTestRouter(t)
	doJSON(t, r, http.MethodPost, "/api/auth/register", "", gin.H{"email": "admin@example.com", "password": "correct horse battery", "fullname": "Admin"})
	token := loginAs(t, r, "admin@example.com", "correct horse battery")

	var admin db.User
	db.DB.Where("user_name = ?", "admin@example.com").First(&admin)

	if w := doJSON(t, r, http.MethodDelete, "/api/user/"+itoa(admin.ID), token, nil); w.Code != http.StatusBadRequest {
		t.Errorf("delete last admin: expected 400, got %d", w.Code)
	}
	if w := doJSON(t, r, http.MethodPut, "/api/user/"+itoa(admin.ID), token, gin.H{"email": "admin@example.com", "fullname": "Admin", "role": "viewer"}); w.Code != http.StatusBadRequest {
		t.Errorf("demote last admin: expected 400, got %d", w.Code)
	}
}

func TestUserByFieldRejectsUnknownColumns(t *testing.T) {
	r := newTestRouter(t)
	doJSON(t, r, http.MethodPost, "/api/auth/register", "", gin.H{"email": "admin@example.com", "password": "correct horse battery", "fullname": "Admin"})
	token := loginAs(t, r, "admin@example.com", "correct horse battery")

	if w := doJSON(t, r, http.MethodGet, "/api/user/field/password/x", token, nil); w.Code != http.StatusBadRequest {
		t.Errorf("unknown field: expected 400, got %d", w.Code)
	}
	if w := doJSON(t, r, http.MethodGet, "/api/user/field/email/admin@example.com", token, nil); w.Code != http.StatusOK {
		t.Errorf("email field: expected 200, got %d", w.Code)
	}
}

func TestDebugRoutesAreGone(t *testing.T) {
	r := newTestRouter(t)
	for _, path := range []string{"/test", "/uix", "/msg/hello", "/chat/hello"} {
		if w := doJSON(t, r, http.MethodGet, path, "", nil); w.Code != http.StatusNotFound {
			t.Errorf("%s: expected 404, got %d", path, w.Code)
		}
	}
}

func TestWebsocketRequiresToken(t *testing.T) {
	r := newTestRouter(t)
	for _, path := range []string{"/ws", "/ws?token=invalid"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: expected 401, got %d", path, w.Code)
		}
	}
}

func TestSecurityHeaders(t *testing.T) {
	r := newTestRouter(t)
	w := doJSON(t, r, http.MethodGet, "/api/auth/register", "", nil)
	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options: got %q", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options: got %q", got)
	}
}

func TestIsAuthorized(t *testing.T) {
	cases := []struct {
		role, method string
		want         bool
	}{
		{db.RoleAdmin, http.MethodGet, true},
		{db.RoleAdmin, http.MethodDelete, true},
		{db.RoleViewer, http.MethodGet, true},
		{db.RoleViewer, http.MethodPost, false},
		{db.RoleViewer, http.MethodPut, false},
		{db.RoleViewer, http.MethodDelete, false},
		{"", http.MethodGet, false},
		{"root", http.MethodGet, false},
	}
	for _, c := range cases {
		if got := IsAuthorized(c.role, c.method); got != c.want {
			t.Errorf("IsAuthorized(%q, %q) = %v, want %v", c.role, c.method, got, c.want)
		}
	}
}

func TestCheckOrigin(t *testing.T) {
	cases := []struct {
		host, origin string
		allowed      []string
		want         bool
	}{
		{"si.example.com:7499", "", nil, true},
		{"si.example.com:7499", "http://si.example.com:7499", nil, true},
		{"si.example.com:7499", "http://evil.example.com", nil, false},
		{"localhost:7499", "http://localhost:8080", []string{"http://localhost:8080"}, true},
		{"localhost:7499", "http://localhost:8081", []string{"http://localhost:8080"}, false},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		req.Host = c.host
		if c.origin != "" {
			req.Header.Set("Origin", c.origin)
		}
		if got := CheckOrigin(req, c.allowed); got != c.want {
			t.Errorf("CheckOrigin(host %q, origin %q) = %v, want %v", c.host, c.origin, got, c.want)
		}
	}
}

func itoa(v uint) string {
	b, _ := json.Marshal(v)
	return string(b)
}
