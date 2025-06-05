// integration_security_test.go
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"moodtracker/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// helper для підготовки інтеграційного сервера
func setupIntegration(t *testing.T) (handler http.Handler, mock sqlmock.Sqlmock, teardown func()) {
	// мок DB
	sqlDB, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("не вдалося створити sqlmock: %v", err)
	}
	db.DB.DB = sqlx.NewDb(sqlDB, "postgres")
	// секрет для JWT
	os.Setenv("JWT_SECRET", "testsecret")

	// збираємо роутер як у main.go
	r := chi.NewRouter()
	r.Route("/auth", RegisterAuthRoutes)
	r.Route("/mood", RegisterMoodRoutes)
	r.Route("/user/telegram", RegisterTelegramRoutes)

	return r, m, func() { sqlDB.Close() }
}

// робимо POST /auth/login, повертаємо токен
func doLogin(t *testing.T, handler http.Handler, mock sqlmock.Sqlmock, email string) string {
	// підготувати очікування DB: спочатку SELECT
	rows := sqlmock.NewRows([]string{"id"}).AddRow("user-1")
	mock.ExpectQuery(`SELECT id FROM users WHERE email=\$1`).
		WithArgs(email).
		WillReturnRows(rows)
	// не вставляємо нового користувача
	// підготуємо запит
	payload := map[string]string{"email": email}
	data, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(data))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Login: очікував 200, отримав %d", rec.Code)
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Login: не вдалося розпарсити JSON: %v", err)
	}
	return resp.Token
}

func Test_Security_Mood_NoToken(t *testing.T) {
	handler, _, teardown := setupIntegration(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodPost, "/mood", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("NoToken: очікував 401, отримав %d", rec.Code)
	}
}

func Test_Security_Mood_InvalidToken(t *testing.T) {
	handler, _, teardown := setupIntegration(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/mood", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("InvalidToken: очікував 401, отримав %d", rec.Code)
	}
}

func Test_Security_Mood_WithValidToken(t *testing.T) {
	handler, mock, teardown := setupIntegration(t)
	defer teardown()

	// 1) логін → отримати токен
	token := doLogin(t, handler, mock, "test@example.com")

	// 2) підготувати мок на вставку mood
	mock.ExpectExec(`INSERT INTO mood`).
		WithArgs(sqlmock.AnyArg(), "user-1", sqlmock.AnyArg(), "🙂", "ok", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// відправляємо запит
	payload := map[string]string{"icon": "🙂", "comment": "ok"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/mood", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("WithValidToken Mood: очікував 201, отримав %d", rec.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mood mock expectations: %v", err)
	}
}

func Test_Security_Telegram_NoToken(t *testing.T) {
	handler, _, teardown := setupIntegration(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodPost, "/user/telegram/register", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Telegram NoToken: очікував 401, отримав %d", rec.Code)
	}
}

func Test_Security_Telegram_WithValidToken(t *testing.T) {
	handler, mock, teardown := setupIntegration(t)
	defer teardown()

	// логін для токена
	token := doLogin(t, handler, mock, "test2@example.com")

	// мок на оновлення telegram_chat_id
	mock.ExpectExec(`UPDATE users SET telegram_chat_id=\$1 WHERE id=\$2`).
		WithArgs(int64(7777), "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// запит
	payload := map[string]int64{"chat_id": 7777}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/user/telegram/register", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("WithValidToken Telegram: очікував 204, отримав %d", rec.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Telegram mock expectations: %v", err)
	}
}
