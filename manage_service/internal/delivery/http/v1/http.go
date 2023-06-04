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
	router.Post("/login-admin", u.LoginAdmin)
	router.Get("/token-user", u.GetTokenUser)
	router.Post("/user", u.CreateUser)
	router.Get("/token", u.IntrospectToken)
	router.Get("/client", u.GetClients)
	router.Post("/login", u.Login)
	router.Post("/login-otp", u.LoginOtp)
	router.Post("/logout", u.Logout)
	router.Post("/login-client", u.LoginClient)
	router.Post("/refresh-token", u.RefreshToken)
	router.Post("/user-info", u.GetUserInfo)
	router.Post("set-password", u.SetPassword)
}

func (u *Http) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.LoginAdmin
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
	w.Header().Set("content-type", "application/json")
	var token dto.IntrospectToken
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resourcePermission, err := u.Usecase.UserKeyclock.IntrospectToken(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resourcePermission)
}

func (u *Http) GetClients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.Client
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	client, err := u.Usecase.UserKeyclock.GetClients(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

func (u *Http) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.Login
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.Usecase.UserKeyclock.Login(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)

}

func (u *Http) LoginOtp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.LoginOTP
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.Usecase.UserKeyclock.LoginOtp(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (u *Http) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.Logout
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := u.Usecase.UserKeyclock.Logout(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Logout user success")
}

func (u *Http) LoginClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.LoginClient
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.Usecase.UserKeyclock.LoginClient(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (u *Http) RefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.RefreshToken
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.Usecase.UserKeyclock.RefreshToken(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (u *Http) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.UserInfo
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userInfo, err := u.Usecase.UserKeyclock.GetUserInfo(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userInfo)
}

func (u *Http) SetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user dto.SetPassword
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := u.Usecase.UserKeyclock.SetPassword(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Set password user success")
}
