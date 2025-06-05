package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"moodtracker/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupAuthTest(t *testing.T) (sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("не вдалося створити sqlmock: %v", err)
	}
	sqlxDB := sqlx.NewDb(mockDB, "postgres")
	db.DB.DB = sqlxDB

	// встановлюємо секрет для підпису JWT
	if err := os.Setenv("JWT_SECRET", "testsecret"); err != nil {
		t.Fatalf("не вдалося встановити JWT_SECRET: %v", err)
	}

	return mock, func() { mockDB.Close() }
}

func TestLoginHandler_BadJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("очікував 400, отримав %d", w.Code)
	}
}

func TestLoginHandler_EmptyEmail(t *testing.T) {
	body, _ := json.Marshal(map[string]string{"email": ""})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("очікував 400, отримав %d", w.Code)
	}
}

func TestLoginHandler_ExistingUser(t *testing.T) {
	mock, teardown := setupAuthTest(t)
	defer teardown()

	// налаштуємо очікування SELECT id
	rows := sqlmock.NewRows([]string{"id"}).AddRow("user-123")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email=$1")).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	// виконуємо запит
	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	LoginHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("очікував 200, отримав %d", w.Code)
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("не вдалося розпарсити JSON: %v", err)
	}
	if resp.Token == "" {
		t.Error("очікував непустий токен")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("не виконані очікування sqlmock: %v", err)
	}
}

func TestLoginHandler_NewUser(t *testing.T) {
	mock, teardown := setupAuthTest(t)
	defer teardown()

	// SELECT повертає ErrNoRows
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM users WHERE email=$1")).
		WithArgs("new@example.com").
		WillReturnError(sql.ErrNoRows)

	// очікуємо INSERT
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(sqlmock.AnyArg(), "new@example.com").
		WillReturnResult(sqlmock.NewResult(1, 1))

	body, _ := json.Marshal(map[string]string{"email": "new@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	LoginHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("очікував 200, отримав %d", w.Code)
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("не вдалося розпарсити JSON: %v", err)
	}
	if resp.Token == "" {
		t.Error("очікував непустий токен")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("не виконані очікування sqlmock: %v", err)
	}
}
