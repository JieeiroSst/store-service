package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/chat-service/dto"
	"github.com/JIeeiroSst/chat-service/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type Http struct {
	Usecase usecase.Usecase
}

func NewHttpV1(Usecase usecase.Usecase) *Http {
	return &Http{
		Usecase: Usecase,
	}
}

func (u *Http) SetupRoutes(router chi.Router) {
	router.Get("/message/:id", u.GetMessageById)
	router.Post("/report", u.CreateReport)
	router.Get("/report/:user-id", u.GetReportByUser)
	router.Delete("/message/:user-id/:message-id", u.DeleteMessage)
}

func (u *Http) GetMessageById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	message, err := u.Usecase.Messages.GetMessageById(r.Context(), id)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	messageJson, err := json.Marshal(&message)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	w.Write(messageJson)
}

func (u *Http) CreateReport(w http.ResponseWriter, r *http.Request) {
	var report dto.Reports
	err := json.NewDecoder(r.Body).Decode(&report)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := u.Usecase.Messages.CreateReport(r.Context(), report); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Person: %+v", report)
}

func (u *Http) GetReportByUser(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(chi.URLParam(r, "user-id"))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	reports, err := u.Usecase.GetReportByUser(r.Context(), userId)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	fmt.Fprintf(w, "Person: %+v", reports)
}

func (u *Http) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(chi.URLParam(r, "user-id"))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	messageId, err := strconv.Atoi(chi.URLParam(r, "message-id"))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	if err := u.Usecase.Messages.DeleteMessage(r.Context(), messageId, userId); err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

}
