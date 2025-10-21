// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Config minima per PKCE
type AuthConfig struct {
	AuthorizationEndpoint string // e.g. https://.../auth
	TokenEndpoint         string // e.g. https://.../token
	ClientID              string
	RedirectURI           string   // e.g. http://localhost:4000/callback
	Scopes                []string // es. ["openid","profile","email"]
	HTTPClient            *http.Client
}

type PKCE struct {
	Verifier  string
	Challenge string
	State     string
}

// NewPKCE genera verifier/challenge/state
func NewPKCE() *PKCE {
	const cs = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	verifier := randomStringCharset(64, cs)
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])
	return &PKCE{
		Verifier:  verifier,
		Challenge: challenge,
		State:     randomStringCharset(32, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"),
	}
}

func randomStringCharset(n int, cs string) string {
	b := make([]byte, n)
	for i := range b {
		_, _ = rand.Read(b[i : i+1])
		b[i] = cs[int(b[i])%len(cs)]
	}
	return string(b)
}

// BuildAuthURL costruisce la URL /authorize identica a quella che avevi prima
func BuildAuthURL(cfg AuthConfig, pkce *PKCE) (string, error) {
	scope := url.QueryEscape(joinScopes(cfg.Scopes))
	v := url.Values{
		"response_type":         {"code"},
		"client_id":             {cfg.ClientID},
		"redirect_uri":          {cfg.RedirectURI},
		"code_challenge":        {pkce.Challenge},
		"code_challenge_method": {"S256"},
		"state":                 {pkce.State},
	}
	if cfg.AuthorizationEndpoint == "" {
		return "", fmt.Errorf("authorization_endpoint is empty")
	}
	return cfg.AuthorizationEndpoint + "?" + v.Encode() + "&scope=" + scope, nil
}

func joinScopes(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	out := scopes[0]
	for i := 1; i < len(scopes); i++ {
		if scopes[i] != "" {
			out += " " + scopes[i]
		}
	}
	return out
}

// OnSuccess viene chiamata dopo lo scambio del code con i token.
// Puoi scrivere direttamente su w (HTML) e fare side-effect (persistenza ecc).
type OnSuccess func(tokens []byte, w http.ResponseWriter)
type OnError func(err error, w http.ResponseWriter)

// StartAuthCodeServer avvia un server HTTP in ascolto su RedirectURI e gestisce la callback.
// - Valida lo state
// - Scambia il code con i token (POST /token) usando PKCE verifier
// - Chiama onSuccess(tokens,w) oppure onError
// Restituisce stop() per spegnere il server.
func StartAuthCodeServer(cfg AuthConfig, pkce *PKCE, onSuccess OnSuccess, onError OnError) (func(), error) {
	u, err := url.Parse(cfg.RedirectURI)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect uri: %w", err)
	}
	host := u.Host // e.g. "localhost:4000"
	path := u.Path // e.g. "/callback"

	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 15 * time.Second}
	}

	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if state != pkce.State {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			if onError != nil {
				onError(fmt.Errorf("state mismatch"), w)
			}
			return
		}
		if code == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			if onError != nil {
				onError(fmt.Errorf("missing authorization code"), w)
			}
			return
		}

		tkn, err := exchangeToken(hc, cfg.TokenEndpoint, cfg.ClientID, pkce.Verifier, code, cfg.RedirectURI)
		if err != nil {
			http.Error(w, "Failed token exchange", http.StatusInternalServerError)
			if onError != nil {
				onError(err, w)
			}
			return
		}
		if onSuccess != nil {
			onSuccess(tkn, w)
		}
	})

	srv := &http.Server{
		Addr:              host,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = srv.Serve(ln)
	}()

	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
		_ = ln.Close()
	}
	return stop, nil
}

func exchangeToken(hc *http.Client, tokenURL, clientID, verifier, code, redirectURI string) ([]byte, error) {
	v := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code_verifier": {verifier},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	resp, err := hc.PostForm(tokenURL, v)
	if err != nil {
		return nil, fmt.Errorf("token request error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		// mantieni stile “status - message” se presente
		var m map[string]any
		msg := resp.Status
		if json.Unmarshal(body, &m) == nil {
			if s, ok := m["message"].(string); ok && s != "" {
				msg = msg + " - " + s
			}
		}
		return nil, fmt.Errorf("token error %s: %s", msg, string(body))
	}
	return body, nil
}
