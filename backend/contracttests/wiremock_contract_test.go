package contracttests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	wiremock "github.com/walkerus/go-wiremock"
)

const (
	baseURL  = "http://localhost:8081"
	adminURL = baseURL + "/__admin/requests"
	authPath = "/auth/login"
	moodPath = "/mood"
	jwtToken = "Bearer abc.def.ghi"
)

type adminRequests struct {
	Requests []struct {
		Request struct {
			Method string `json:"method"`
			URL    string `json:"url"`
		} `json:"request"`
	} `json:"requests"`
}

// countRequests звертається до /__admin/requests і рахує запити з даним методом і шляхом
func countRequests(t *testing.T, method, urlPath string) int {
	resp, err := http.Get(adminURL)
	if err != nil {
		t.Fatalf("Не вдалося GET %s: %v", adminURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Admin API повернув %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)

	var ar adminRequests
	if err := json.Unmarshal(body, &ar); err != nil {
		t.Fatalf("Не вдалося розпарсити admin requests: %v", err)
	}
	cnt := 0
	for _, r := range ar.Requests {
		if r.Request.Method == method && r.Request.URL == urlPath {
			cnt++
		}
	}
	return cnt
}

func TestAuthLoginContract(t *testing.T) {
	client := wiremock.NewClient(baseURL)
	defer client.Reset()

	// 1) Реєструємо stub
	err := client.StubFor(
		wiremock.Post(wiremock.URLPathEqualTo(authPath)).
			WithHeader("Content-Type", wiremock.EqualTo("application/json")).
			WithBodyPattern(wiremock.EqualToJson(`{"email":"user@example.com"}`, wiremock.IgnoreArrayOrder, wiremock.IgnoreExtraElements)).
			WillReturnResponse(
				wiremock.NewResponse().
					WithStatus(200).
					WithHeader("Content-Type", "application/json").
					WithBody(`{"token":"jwt.token.here"}`),
			),
	)
	if err != nil {
		t.Fatalf("Не вдалося зареєструвати stub для %s: %v", authPath, err)
	}

	// 2) Виконуємо запит як consumer
	resp, err := http.Post(baseURL+authPath, "application/json",
		bytes.NewReader([]byte(`{"email":"user@example.com"}`)),
	)
	if err != nil {
		t.Fatalf("Помилка POST %s: %v", authPath, err)
	}
	resp.Body.Close()

	// 3) Перевіряємо, що WireMock отримав рівно 1 виклик
	if got := countRequests(t, "POST", authPath); got != 1 {
		t.Errorf("Очікували 1 запит POST %s, отримали %d", authPath, got)
	}
}

func TestCreateMoodContract(t *testing.T) {
	client := wiremock.NewClient(baseURL)
	defer client.Reset()

	payload := map[string]string{"icon": "😊", "comment": "feeling good"}
	payloadBytes, _ := json.Marshal(payload)

	// 1) Реєструємо stub
	err := client.StubFor(
		wiremock.Post(wiremock.URLPathEqualTo(moodPath)).
			WithHeader("Content-Type", wiremock.EqualTo("application/json")).
			WithHeader("Authorization", wiremock.EqualTo(jwtToken)).
			WithBodyPattern(wiremock.EqualToJson(string(payloadBytes), wiremock.IgnoreArrayOrder, wiremock.IgnoreExtraElements)).
			WillReturnResponse(
				wiremock.NewResponse().
					WithStatus(201).
					WithHeader("Content-Type", "application/json").
					WithBody(`{
						"id":"uuid-1234",
						"user_id":"user-1",
						"date":"2025-05-28T12:00:00Z",
						"icon":"😊",
						"comment":"feeling good",
						"created_at":"2025-05-28T12:00:00Z",
						"updated_at":"2025-05-28T12:00:00Z"
					}`),
			),
	)
	if err != nil {
		t.Fatalf("Не вдалося зареєструвати stub для %s: %v", moodPath, err)
	}

	// 2) Виконуємо запит як consumer
	req, _ := http.NewRequest("POST", baseURL+moodPath, bytes.NewReader(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", jwtToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Помилка POST %s: %v", moodPath, err)
	}
	resp.Body.Close()

	// 3) Перевіряємо, що WireMock отримав рівно 1 виклик
	if got := countRequests(t, "POST", moodPath); got != 1 {
		t.Errorf("Очікували 1 запит POST %s, отримали %d", moodPath, got)
	}
}
