package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/chat-service/dto"
	"github.com/JIeeiroSst/chat-service/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type stubMessages struct {
	gotID, gotUser, gotMessage int
}

func (s *stubMessages) SaveMessage(context.Context, dto.Messages) error { return nil }
func (s *stubMessages) CreateReport(context.Context, dto.Reports) error { return nil }

func (s *stubMessages) GetMessageById(_ context.Context, id int) (*dto.Messages, error) {
	s.gotID = id
	return &dto.Messages{ID: id, MessageType: "text"}, nil
}

func (s *stubMessages) GetReportByUser(_ context.Context, userID int) ([]dto.Reports, error) {
	s.gotUser = userID
	return []dto.Reports{{ID: 1, UserId: userID, Status: "open"}}, nil
}

func (s *stubMessages) DeleteMessage(_ context.Context, messageID, userID int) error {
	s.gotMessage, s.gotUser = messageID, userID
	return nil
}

func newRouter(s *stubMessages) http.Handler {
	r := chi.NewRouter()
	NewHttpV1(&usecase.Usecase{Messages: s}).SetupRoutes(r)
	return r
}

func TestRoutesMatchRealIDs(t *testing.T) {
	s := &stubMessages{}
	r := newRouter(s)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/message/42", nil))
	var msg dto.Messages
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &msg) != nil || msg.ID != 42 || s.gotID != 42 {
		t.Fatalf("GET /message/42 = %d %s", rec.Code, rec.Body)
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/report/7", nil))
	var reports []dto.Reports
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" ||
		json.Unmarshal(rec.Body.Bytes(), &reports) != nil || len(reports) != 1 || reports[0].UserId != 7 {
		t.Fatalf("GET /report/7 = %d %s", rec.Code, rec.Body)
	}

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/message/7/42", nil))
	if rec.Code != http.StatusOK || s.gotUser != 7 || s.gotMessage != 42 {
		t.Fatalf("DELETE /message/7/42 = %d user=%d message=%d", rec.Code, s.gotUser, s.gotMessage)
	}
}
