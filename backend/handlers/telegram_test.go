package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"moodtracker/db"
	"moodtracker/middleware"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupTelegramTest(t *testing.T) (mock sqlmock.Sqlmock, teardown func()) {
	sqlDB, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("не вдалося створити sqlmock: %v", err)
	}
	db.DB.DB = sqlx.NewDb(sqlDB, "postgres")
	return m, func() { sqlDB.Close() }
}

func TestRegisterTelegram_BadJSON(t *testing.T) {
	// некоректний JSON
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not-json"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	w := httptest.NewRecorder()

	RegisterTelegram(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("очікував 400 Bad Request, отримав %d", w.Code)
	}
}

func TestRegisterTelegram_DBError(t *testing.T) {
	mock, teardown := setupTelegramTest(t)
	defer teardown()

	// імітуємо помилку оновлення в БД
	mock.
		ExpectExec(regexp.QuoteMeta("UPDATE users SET telegram_chat_id=$1 WHERE id=$2")).
		WithArgs(int64(1234), "user-1").
		WillReturnError(errors.New("update fail"))

	// формуємо валідний запит
	payload := chatReq{ChatID: 1234}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	w := httptest.NewRecorder()

	RegisterTelegram(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("очікував 500 Internal Server Error, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("невиконані очікування sqlmock: %v", err)
	}
}

func TestRegisterTelegram_Success(t *testing.T) {
	mock, teardown := setupTelegramTest(t)
	defer teardown()

	// імітуємо успішне оновлення
	mock.
		ExpectExec(regexp.QuoteMeta("UPDATE users SET telegram_chat_id=$1 WHERE id=$2")).
		WithArgs(int64(5678), "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := chatReq{ChatID: 5678}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	w := httptest.NewRecorder()

	RegisterTelegram(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("очікував 204 No Content, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("невиконані очікування sqlmock: %v", err)
	}
}
