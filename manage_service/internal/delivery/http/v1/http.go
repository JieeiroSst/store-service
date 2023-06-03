package v1

import (
	"encoding/json"
	"net/http"

	"github.com/JIeeiroSst/manage-service/internal/dto"
	"github.com/JIeeiroSst/manage-service/internal/usecase"
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

}

func (u *Http) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.Login
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.Usecase.UserKeyclock.LoginAdmin(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (u *Http) GetTokenUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	realm := chi.URLParam(r, "realm")
	tokenInfo, err := u.Usecase.UserKeyclock.GetTokenUser(r.Context(), realm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tokenInfo)
}

func (u *Http) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.CreateUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := u.Usecase.UserKeyclock.CreateUser(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (u *Http) IntrospectToken(w http.ResponseWriter, r *http.Request) {
	var token dto.IntrospectToken
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (u *Http) GetClients(w http.ResponseWriter, r *http.Request) {}

func (u *Http) Login(w http.ResponseWriter, r *http.Request) {}

func (u *Http) LoginOtp(w http.ResponseWriter, r *http.Request) {}

func (u *Http) Logout(w http.ResponseWriter, r *http.Request) {}

func (u *Http) LoginClient(w http.ResponseWriter, r *http.Request) {}

func (u *Http) RefreshToken(w http.ResponseWriter, r *http.Request) {}

func (u *Http) GetUserInfo(w http.ResponseWriter, r *http.Request) {}

func (u *Http) SetPassword(w http.ResponseWriter, r *http.Request) {}
