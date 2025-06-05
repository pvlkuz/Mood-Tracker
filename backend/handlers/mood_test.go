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
	"time"

	"moodtracker/db"
	"moodtracker/middleware"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// setupMoodTest створює sqlmock і підставляє його в db.DB
func setupMoodTest(t *testing.T) (mock sqlmock.Sqlmock, teardown func()) {
	sqlDB, m, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("не вдалося створити sqlmock: %v", err)
	}
	db.DB.DB = sqlx.NewDb(sqlDB, "postgres")
	return m, func() { sqlDB.Close() }
}

// newRequest формує http.Request з контекстом userID і, за потреби, chi URLParam "id"
func newRequest(method, url string, body []byte, id string) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	// додаємо userID
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "user-1")
	// якщо потрібен параметр {id}
	if id != "" {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id)
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	}
	return req.WithContext(ctx)
}

func TestCreateMood_BadJSON(t *testing.T) {
	req := newRequest(http.MethodPost, "/mood", []byte("not-json"), "")
	w := httptest.NewRecorder()

	CreateMood(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("TestCreateMood_BadJSON: очікував 400, отримав %d", w.Code)
	}
}

func TestCreateMood_DBError(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	// будь-який INSERT повертає помилку
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO mood")).
		WillReturnError(errors.New("insert fail"))

	payload := map[string]string{"icon": "😀", "comment": "oops"}
	body, _ := json.Marshal(payload)
	req := newRequest(http.MethodPost, "/mood", body, "")
	w := httptest.NewRecorder()

	CreateMood(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("TestCreateMood_DBError: очікував 500, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreateMood_DBError: невиконані очікування: %v", err)
	}
}

func TestCreateMood_Success(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO mood")).
		WithArgs(sqlmock.AnyArg(), "user-1", sqlmock.AnyArg(), "😃", "ok", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := map[string]string{"icon": "😃", "comment": "ok"}
	body, _ := json.Marshal(payload)
	req := newRequest(http.MethodPost, "/mood", body, "")
	w := httptest.NewRecorder()

	CreateMood(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("TestCreateMood_Success: очікував 201, отримав %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("TestCreateMood_Success: не вдалося розпарсити JSON: %v", err)
	}
	if resp["user_id"] != "user-1" || resp["icon"] != "😃" || resp["comment"] != "ok" {
		t.Errorf("TestCreateMood_Success: невірні дані у відповіді: %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreateMood_Success: невиконані очікування: %v", err)
	}
}

func TestListMood_DBError(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM mood WHERE user_id=$1")).
		WithArgs("user-1").
		WillReturnError(errors.New("select fail"))

	req := newRequest(http.MethodGet, "/mood", nil, "")
	w := httptest.NewRecorder()

	ListMood(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("TestListMood_DBError: очікував 500, отримав %d", w.Code)
	}
}

func TestListMood_NoFilter_Success(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	now := time.Now().Truncate(time.Second)
	// повертаємо time.Time для date
	rows := sqlmock.NewRows([]string{"id", "user_id", "date", "icon", "comment", "created_at", "updated_at"}).
		AddRow("m1", "user-1", now, "🙂", "fine", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM mood WHERE user_id=$1")).
		WithArgs("user-1").
		WillReturnRows(rows)

	req := newRequest(http.MethodGet, "/mood", nil, "")
	w := httptest.NewRecorder()

	ListMood(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("TestListMood_NoFilter_Success: очікував 200, отримав %d", w.Code)
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("TestListMood_NoFilter_Success: не вдалося розпарсити JSON: %v", err)
	}
	if len(list) != 1 || list[0]["icon"] != "🙂" {
		t.Errorf("TestListMood_NoFilter_Success: неправильні записи: %+v", list)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestListMood_NoFilter_Success: невиконані очікування: %v", err)
	}
}

func TestListMood_WithFilter_Success(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{"id", "user_id", "date", "icon", "comment", "created_at", "updated_at"}).
		AddRow("m2", "user-1", now, "🙁", "sad", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM mood WHERE user_id=$1 AND date BETWEEN $2 AND $3")).
		WithArgs("user-1", "2025-01-01", "2025-01-31").
		WillReturnRows(rows)

	req := newRequest(http.MethodGet, "/mood?from=2025-01-01&to=2025-01-31", nil, "")
	w := httptest.NewRecorder()

	ListMood(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("TestListMood_WithFilter_Success: очікував 200, отримав %d", w.Code)
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("TestListMood_WithFilter_Success: не вдалося розпарсити JSON: %v", err)
	}
	if len(list) != 1 || list[0]["icon"] != "🙁" {
		t.Errorf("TestListMood_WithFilter_Success: неправильні записи: %+v", list)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestListMood_WithFilter_Success: невиконані очікування: %v", err)
	}
}

func TestUpdateMood_BadJSON(t *testing.T) {
	req := newRequest(http.MethodPut, "/mood/m1", []byte("bad"), "m1")
	w := httptest.NewRecorder()

	UpdateMood(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("TestUpdateMood_BadJSON: очікував 400, отримав %d", w.Code)
	}
}

func TestUpdateMood_DBError(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE mood SET icon=$1, comment=$2, updated_at=$3 WHERE id=$4 AND user_id=$5")).
		WithArgs("ico2", "comm2", sqlmock.AnyArg(), "m1", "user-1").
		WillReturnError(errors.New("update fail"))

	payload := map[string]string{"icon": "ico2", "comment": "comm2"}
	body, _ := json.Marshal(payload)
	req := newRequest(http.MethodPut, "/mood/m1", body, "m1")
	w := httptest.NewRecorder()

	UpdateMood(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("TestUpdateMood_DBError: очікував 500, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestUpdateMood_DBError: невиконані очікування: %v", err)
	}
}

func TestUpdateMood_NotFound(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE mood SET icon=$1, comment=$2, updated_at=$3 WHERE id=$4 AND user_id=$5")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	payload := map[string]string{"icon": "ico3", "comment": "comm3"}
	body, _ := json.Marshal(payload)
	req := newRequest(http.MethodPut, "/mood/m1", body, "m1")
	w := httptest.NewRecorder()

	UpdateMood(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("TestUpdateMood_NotFound: очікував 404, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestUpdateMood_NotFound: невиконані очікування: %v", err)
	}
}

func TestUpdateMood_Success(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE mood SET icon=$1, comment=$2, updated_at=$3 WHERE id=$4 AND user_id=$5")).
		WithArgs("ico4", "comm4", sqlmock.AnyArg(), "m1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	payload := map[string]string{"icon": "ico4", "comment": "comm4"}
	body, _ := json.Marshal(payload)
	req := newRequest(http.MethodPut, "/mood/m1", body, "m1")
	w := httptest.NewRecorder()

	UpdateMood(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("TestUpdateMood_Success: очікував 204, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestUpdateMood_Success: невиконані очікування: %v", err)
	}
}

func TestDeleteMood_DBError(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM mood WHERE id=$1 AND user_id=$2")).
		WithArgs("m1", "user-1").
		WillReturnError(errors.New("delete fail"))

	req := newRequest(http.MethodDelete, "/mood/m1", nil, "m1")
	w := httptest.NewRecorder()

	DeleteMood(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("TestDeleteMood_DBError: очікував 500, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestDeleteMood_DBError: невиконані очікування: %v", err)
	}
}

func TestDeleteMood_NotFound(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM mood WHERE id=$1 AND user_id=$2")).
		WithArgs("m1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	req := newRequest(http.MethodDelete, "/mood/m1", nil, "m1")
	w := httptest.NewRecorder()

	DeleteMood(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("TestDeleteMood_NotFound: очікував 404, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestDeleteMood_NotFound: невиконані очікування: %v", err)
	}
}

func TestDeleteMood_Success(t *testing.T) {
	mock, teardown := setupMoodTest(t)
	defer teardown()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM mood WHERE id=$1 AND user_id=$2")).
		WithArgs("m1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := newRequest(http.MethodDelete, "/mood/m1", nil, "m1")
	w := httptest.NewRecorder()

	DeleteMood(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("TestDeleteMood_Success: очікував 204, отримав %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestDeleteMood_Success: невиконані очікування: %v", err)
	}
}
