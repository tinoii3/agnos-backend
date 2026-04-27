package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"agnos-backend/internal/auth"
	"agnos-backend/internal/store"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type mockStore struct {
	createStaffFunc        func(ctx context.Context, username, hashedPassword, hospital string) error
	getStaffByUsernameFunc func(ctx context.Context, username string) (store.Staff, error)
	searchPatientsFunc     func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error)
}

func (m *mockStore) CreateStaff(ctx context.Context, username, hashedPassword, hospital string) error {
	return m.createStaffFunc(ctx, username, hashedPassword, hospital)
}

func (m *mockStore) GetStaffByUsername(ctx context.Context, username string) (store.Staff, error) {
	return m.getStaffByUsernameFunc(ctx, username)
}

func (m *mockStore) SearchPatients(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
	return m.searchPatientsFunc(ctx, hospital, filters)
}

func TestCreateStaffPositive(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			if username != "alice" || hospital != "hospital-a" {
				t.Fatalf("unexpected payload: %s %s", username, hospital)
			}
			if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("password123")); err != nil {
				t.Fatalf("password not hashed correctly: %v", err)
			}
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/create", bytes.NewBufferString(`{"username":"alice","password":"password123","hospital":"hospital-a"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCreateStaffNegative(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/create", bytes.NewBufferString(`{"username":"alice"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateStaffNegativeDuplicateUsername(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return &pgconn.PgError{Code: "23505"}
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/create", bytes.NewBufferString(`{"username":"alice","password":"password123","hospital":"hospital-a"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestCreateStaffNegativeStoreError(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return errors.New("db down")
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/create", bytes.NewBufferString(`{"username":"alice","password":"password123","hospital":"hospital-a"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestLoginPositive(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{
				Username:       "alice",
				HashedPassword: string(hash),
				Hospital:       "hospital-a",
			}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/login", bytes.NewBufferString(`{"username":"alice","password":"password123","hospital":"hospital-a"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if body["token"] == "" {
		t.Fatalf("expected token in response")
	}
}

func TestLoginNegative(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, pgx.ErrNoRows
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/login", bytes.NewBufferString(`{"username":"alice","password":"wrong","hospital":"hospital-a"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginNegativeHospitalMismatch(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{
				Username:       "alice",
				HashedPassword: string(hash),
				Hospital:       "hospital-a",
			}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/login", bytes.NewBufferString(`{"username":"alice","password":"password123","hospital":"hospital-b"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginNegativeStoreError(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, errors.New("db timeout")
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/staff/login", bytes.NewBufferString(`{"username":"alice","password":"password123","hospital":"hospital-a"}`))
	req.Header.Set("Content-Type", "application/json")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestPatientSearchPositive(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			if hospital != "hospital-a" {
				t.Fatalf("expected hospital-a, got %s", hospital)
			}
			if filters.FirstName != "Somchai" || filters.NationalID != "1100100100001" {
				t.Fatalf("unexpected filters: %+v", filters)
			}
			return []store.Patient{
				{
					ID:           1,
					FirstNameTH:  "สมชาย",
					MiddleNameTH: "",
					LastNameTH:   "ใจดี",
					FirstNameEN:  "Somchai",
					MiddleNameEN: "",
					LastNameEN:   "Jaidee",
					DateOfBirth:  "1990-01-15",
					PatientHN:    "HN-A-001",
					NationID:     "1100100100001",
					PassportID:   "",
					PhoneNumber:  "0812345671",
					Email:        "somchai@email.com",
					Gender:       "M",
					Hospital:     "hospital-a",
				},
			}, nil
		},
	}
	server := NewServer(st, "secret")

	token, _ := auth.GenerateToken("secret", "alice", "hospital-a")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/patient/search?first_name=Somchai&national_id=1100100100001", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Patients []store.Patient `json:"patients"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Patients) != 1 || body.Patients[0].FirstNameEN != "Somchai" {
		t.Fatalf("unexpected response payload: %s", w.Body.String())
	}
}

func TestPatientSearchNegative(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/patient/search", nil)
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPatientSearchNegativeInvalidAuthorizationHeader(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/patient/search?first_name=Somchai", nil)
	req.Header.Set("Authorization", "Token abc")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPatientSearchNegativeInvalidToken(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, nil
		},
	}
	server := NewServer(st, "secret")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/patient/search?first_name=Somchai", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-token")
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestPatientSearchNegativeStoreError(t *testing.T) {
	st := &mockStore{
		createStaffFunc: func(ctx context.Context, username, hashedPassword, hospital string) error {
			return nil
		},
		getStaffByUsernameFunc: func(ctx context.Context, username string) (store.Staff, error) {
			return store.Staff{}, nil
		},
		searchPatientsFunc: func(ctx context.Context, hospital string, filters store.PatientFilters) ([]store.Patient, error) {
			return nil, context.DeadlineExceeded
		},
	}
	server := NewServer(st, "secret")

	token, _ := auth.GenerateToken("secret", "alice", "hospital-a")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/patient/search?first_name=Somchai", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
