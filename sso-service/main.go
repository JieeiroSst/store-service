package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var cfg = &oauth2.Config{
	ClientID:     "client_id",
	ClientSecret: "client_secret",
	RedirectURL:  "http://localhost:8080/callback",
	Scopes:       []string{"openid", "profile", "email"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "http://localhost:8080/authorize",
		TokenURL: "http://localhost:8080/token",
	},
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	url := cfg.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "access_token",
		Value: token.AccessToken,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func userHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := r.URL.Query().Get("access_token")
	if accessToken == "" {
		http.Error(w, "Missing access token", http.StatusBadRequest)
		return
	}

	claims, err := verifyJWT(accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf("Hello, %+v!", claims)))
}

func verifyJWT(accessToken string) (*oidc.IDTokenVerifier, error) {
	provider, err := oidc.NewProvider(context.Background(), "http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	idToken, err := provider.Verifier(&oidc.Config{ClientID: "client_id"}).Verify(context.Background(), accessToken)
	if err != nil {
		return nil, err
	}

	var claims oidc.IDTokenVerifier
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/user", userHandler)
	log.Fatal(http.ListenAndServe(":8090", nil))
}
