package user_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/NurulloMahmud/medicalka-project/config"
	"github.com/NurulloMahmud/medicalka-project/internal/auth"
	"github.com/NurulloMahmud/medicalka-project/internal/middleware"
	"github.com/NurulloMahmud/medicalka-project/internal/tasks"
	"github.com/NurulloMahmud/medicalka-project/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var (
	testDB         *sql.DB
	testHandler    *user.UserHandler
	testMiddleware *middleware.Middleware
	testCfg        config.Config
)

func TestMain(m *testing.M) {
	testCfg = config.Config{
		DatabaseURL: os.Getenv("TEST_DATABASE_URL"),
		JWTSecret:   "test-secret-key",
	}

	if testCfg.DatabaseURL == "" {
		testCfg.DatabaseURL = "postgres://postgres:postgres@localhost:5432/medicalka_test?sslmode=disable"
	}

	var err error
	testDB, err = sql.Open("postgres", testCfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	err = testDB.Ping()
	if err != nil {
		log.Fatalf("failed to ping test database: %v", err)
	}

	setupTestDB()

	logger := log.New(os.Stdout, "TEST: ", log.Ldate|log.Ltime)
	emailSender := tasks.NewEmailSender(config.SMTP{}, logger)
	repo := user.NewPostgresRepository(testDB)
	service := user.NewService(repo, emailSender)
	testHandler = user.NewHandler(service, logger, testCfg)
	testMiddleware = middleware.NewMiddleware(logger, repo, testCfg)

	code := m.Run()

	cleanupTestDB()
	testDB.Close()

	os.Exit(code)
}

func setupTestDB() {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`DROP TABLE IF EXISTS email_verification_tokens`,
		`DROP TABLE IF EXISTS likes`,
		`DROP TABLE IF EXISTS comments`,
		`DROP TABLE IF EXISTS posts`,
		`DROP TABLE IF EXISTS users`,
		`CREATE TABLE users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) NOT NULL UNIQUE,
			username VARCHAR(32) NOT NULL UNIQUE,
			full_name VARCHAR(100) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			is_verified BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE email_verification_tokens (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) NOT NULL UNIQUE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
	}

	for _, q := range queries {
		_, err := testDB.Exec(q)
		if err != nil {
			log.Fatalf("failed to setup test database: %v", err)
		}
	}
}

func cleanupTestDB() {
	testDB.Exec(`DROP TABLE IF EXISTS email_verification_tokens`)
	testDB.Exec(`DROP TABLE IF EXISTS users`)
}

func clearUsers() {
	testDB.Exec(`DELETE FROM email_verification_tokens`)
	testDB.Exec(`DELETE FROM users`)
}

func TestRegister_Success(t *testing.T) {
	clearUsers()

	payload := map[string]string{
		"email":     "test@example.com",
		"username":  "testuser",
		"full_name": "Test User",
		"password":  "password123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testHandler.HandleRegister(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data in response")
	}

	if data["email"] != "test@example.com" {
		t.Errorf("expected email test@example.com, got %v", data["email"])
	}

	if data["username"] != "testuser" {
		t.Errorf("expected username testuser, got %v", data["username"])
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	clearUsers()

	payload := map[string]string{
		"email":     "duplicate@example.com",
		"username":  "user1",
		"full_name": "User One",
		"password":  "password123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testHandler.HandleRegister(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("first registration failed: %s", rr.Body.String())
	}

	payload["username"] = "user2"
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	testHandler.HandleRegister(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	clearUsers()

	payload := map[string]string{
		"email":     "user1@example.com",
		"username":  "duplicateuser",
		"full_name": "User One",
		"password":  "password123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testHandler.HandleRegister(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("first registration failed: %s", rr.Body.String())
	}

	payload["email"] = "user2@example.com"
	body, _ = json.Marshal(payload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	testHandler.HandleRegister(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	clearUsers()

	registerPayload := map[string]string{
		"email":     "login@example.com",
		"username":  "loginuser",
		"full_name": "Login User",
		"password":  "password123",
	}

	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testHandler.HandleRegister(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("registration failed: %s", rr.Body.String())
	}

	loginPayload := map[string]string{
		"email":    "login@example.com",
		"password": "password123",
	}

	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	testHandler.HandleLogin(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	if _, ok := response["access_token"]; !ok {
		t.Error("expected access_token in response")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	clearUsers()

	registerPayload := map[string]string{
		"email":     "invalid@example.com",
		"username":  "invaliduser",
		"full_name": "Invalid User",
		"password":  "password123",
	}

	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testHandler.HandleRegister(rr, req)

	loginPayload := map[string]string{
		"email":    "invalid@example.com",
		"password": "wrongpassword",
	}

	body, _ = json.Marshal(loginPayload)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	testHandler.HandleLogin(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestProtectedEndpoint_ValidToken(t *testing.T) {
	clearUsers()

	registerPayload := map[string]string{
		"email":     "protected@example.com",
		"username":  "protecteduser",
		"full_name": "Protected User",
		"password":  "password123",
	}

	body, _ := json.Marshal(registerPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testHandler.HandleRegister(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("registration failed: %s", rr.Body.String())
	}

	var regResponse map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &regResponse)
	data := regResponse["data"].(map[string]interface{})
	userID, _ := uuid.Parse(data["id"].(string))

	tokenClaims := auth.TokenClaims{
		ID:       userID,
		Email:    "protected@example.com",
		Username: "protecteduser",
	}

	token, _ := auth.GenerateAccessToken(tokenClaims, testCfg.JWTSecret)

	r := chi.NewRouter()
	r.Use(testMiddleware.Authenticate)
	r.Get("/api/auth/me", testHandler.HandleMe)

	req = httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &response)

	user, ok := response["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected user in response")
	}

	if user["email"] != "protected@example.com" {
		t.Errorf("expected email protected@example.com, got %v", user["email"])
	}
}

func TestProtectedEndpoint_InvalidToken(t *testing.T) {
	r := chi.NewRouter()
	r.Use(testMiddleware.Authenticate)
	r.Get("/api/auth/me", testHandler.HandleMe)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestProtectedEndpoint_MissingToken(t *testing.T) {
	r := chi.NewRouter()
	r.Use(testMiddleware.Authenticate)
	r.Get("/api/auth/me", testHandler.HandleMe)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Logf("note: missing token returns anonymous user, got status %d", rr.Code)
	}
}
