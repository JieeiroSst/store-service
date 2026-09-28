package http

import (
	"encoding/json"
	"net/http"

	"github.com/JIeeiroSst/manage-service/internal/domain/model"
	"github.com/go-chi/chi/v5"
)

func (u *Handler) authRoutes(router chi.Router) {
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
	router.Post("/set-password", u.SetPassword)
}

func (u *Handler) LoginAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.LoginAdmin
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.auth.LoginAdmin(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (u *Handler) GetTokenUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	realm := r.URL.Query().Get("realm")
	tokenInfo, err := u.auth.GetTokenUser(r.Context(), realm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tokenInfo)
}

func (u *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.CreateUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := u.auth.CreateUser(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (u *Handler) IntrospectToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var token model.IntrospectToken
	if err := json.NewDecoder(r.Body).Decode(&token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resourcePermission, err := u.auth.IntrospectToken(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resourcePermission)
}

func (u *Handler) GetClients(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.Client
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	client, err := u.auth.GetClients(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

func (u *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.Login
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.auth.Login(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)

}

func (u *Handler) LoginOtp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.LoginOTP
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.auth.LoginOtp(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (u *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.Logout
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := u.auth.Logout(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Logout user success")
}

func (u *Handler) LoginClient(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.LoginClient
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.auth.LoginClient(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (u *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.RefreshToken
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, err := u.auth.RefreshToken(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(token)
}

func (u *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.UserInfo
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userInfo, err := u.auth.GetUserInfo(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userInfo)
}

func (u *Handler) SetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var user model.SetPassword
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := u.auth.SetPassword(r.Context(), user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Set password user success")
}
