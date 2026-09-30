package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"WorldOfWarcraft/internal/store"
)

func newTestServer() *Server {
	return New(store.NewMemory())
}

func do(t *testing.T, srv *Server, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := do(t, newTestServer(), http.MethodGet, "/health", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestCreateCharacter(t *testing.T) {
	rec := do(t, newTestServer(), http.MethodPost, "/characters",
		`{"name":"thrall","race":"orc","class":"shaman"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", rec.Code, rec.Body)
	}
	if got := rec.Header().Get("Location"); got != "/characters/thrall" {
		t.Errorf("Location = %q, want %q", got, "/characters/thrall")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestCreateCharacterErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
		want int
	}{
		{"invalid combo", `{"name":"sylvanas","race":"orc","class":"druid"}`, http.StatusBadRequest},
		{"unknown race", `{"name":"gizmo","race":"gnome","class":"mage"}`, http.StatusBadRequest},
		{"unknown class", `{"name":"gizmo","race":"orc","class":"warlock"}`, http.StatusBadRequest},
		{"empty name", `{"name":"","race":"orc","class":"shaman"}`, http.StatusBadRequest},
		{"unknown field", `{"nickname":"thrall","race":"orc","class":"shaman"}`, http.StatusBadRequest},
		{"malformed json", `{"name":`, http.StatusBadRequest},
		{"empty body", ``, http.StatusBadRequest},
		{"trailing garbage", `{"name":"a","race":"orc","class":"shaman"}{"x":1}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, newTestServer(), http.MethodPost, "/characters", tt.body)
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d; body = %s", rec.Code, tt.want, rec.Body)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
		})
	}
}

func TestCreateCharacterDuplicate(t *testing.T) {
	srv := newTestServer()
	body := `{"name":"Thrall","race":"orc","class":"shaman"}`

	if rec := do(t, srv, http.MethodPost, "/characters", body); rec.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, want 201", rec.Code)
	}
	rec := do(t, srv, http.MethodPost, "/characters", body)
	if rec.Code != http.StatusConflict {
		t.Errorf("duplicate status = %d, want 409; body = %s", rec.Code, rec.Body)
	}
}

func TestGetCharacter(t *testing.T) {
	srv := newTestServer()
	do(t, srv, http.MethodPost, "/characters", `{"name":"Thrall","race":"orc","class":"shaman"}`)

	rec := do(t, srv, http.MethodGet, "/characters/thrall", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body)
	}
}

func TestGetCharacterNotFound(t *testing.T) {
	rec := do(t, newTestServer(), http.MethodGet, "/characters/nobody", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestListCharacters(t *testing.T) {
	srv := newTestServer()
	empty := do(t, srv, http.MethodGet, "/characters", "")
	if empty.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", empty.Code)
	}
	if got := strings.TrimSpace(empty.Body.String()); got != "[]" {
		t.Errorf("empty list body = %q, want []", got)
	}

	do(t, srv, http.MethodPost, "/characters", `{"name":"thrall","race":"orc","class":"shaman"}`)
	do(t, srv, http.MethodPost, "/characters", `{"name":"garona","race":"orc","class":"warrior"}`)

	rec := do(t, srv, http.MethodGet, "/characters", "")
	if !strings.Contains(rec.Body.String(), "garona") || !strings.Contains(rec.Body.String(), "thrall") {
		t.Errorf("list missing characters: %s", rec.Body)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec := do(t, newTestServer(), http.MethodDelete, "/characters", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestStatusForUnmappedErrorIsInternal(t *testing.T) {
	if got := statusFor(errNotMapped{}); got != http.StatusInternalServerError {
		t.Errorf("statusFor(unmapped) = %d, want 500", got)
	}
}

type errNotMapped struct{}

func (errNotMapped) Error() string { return "unmapped" }
