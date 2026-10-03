package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/chat-service/dto"
	"github.com/JIeeiroSst/chat-service/internal/usecase"
	"github.com/go-chi/chi/v5"
)

type Http struct {
	Usecase *usecase.Usecase
}

func NewHttpV1(Usecase *usecase.Usecase) *Http {
	return &Http{
		Usecase: Usecase,
	}
}

func (u *Http) SetupRoutes(router chi.Router) {
	router.Get("/message/{id}", u.GetMessageById)
	router.Post("/report", u.CreateReport)
	router.Get("/report/{userID}", u.GetReportByUser)
	router.Delete("/message/{userID}/{messageID}", u.DeleteMessage)
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
	writeJSON(w, http.StatusOK, message)
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
	writeJSON(w, http.StatusCreated, report)
}

func (u *Http) GetReportByUser(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	reports, err := u.Usecase.GetReportByUser(r.Context(), userId)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	writeJSON(w, http.StatusOK, reports)
}

func (u *Http) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	messageId, err := strconv.Atoi(chi.URLParam(r, "messageID"))
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}
	if err := u.Usecase.Messages.DeleteMessage(r.Context(), messageId, userId); err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
