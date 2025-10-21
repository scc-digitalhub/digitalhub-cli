// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package old

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"

	"dhcli/utils"
)

const redirectURI = "http://localhost:4000/callback"

var generatedState string

// Runs PKCE flow for authentication
func LoginHandler() error {
	// Ensure environment is up-to-date and compatible
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.LoginMin, utils.LoginMax)

	// PKCE
	verifier, challenge := generatePKCE()
	generatedState = randomString(32)

	// Start local callback server
	stop, err := startAuthCodeServer(verifier)
	if err != nil {
		return fmt.Errorf("impossibile avviare il server locale: %w", err)
	}
	defer stop()

	// Build authorize URL
	authURL, err := buildAuthURL(challenge, generatedState)
	if err != nil {
		return err
	}

	fmt.Println("─────────────────────────────────────────────────────────────────────")
	fmt.Println("  The following URL will be opened in your browser to authenticate:  ")
	fmt.Println("─────────────────────────────────────────────────────────────────────")
	fmt.Println(authURL)
	fmt.Println("─────────────────────────────────────────────────────────────────────")
	fmt.Print("Press Enter to continue... ")

	if _, err := bufio.NewReader(os.Stdin).ReadBytes('\n'); err != nil {
		return fmt.Errorf("errore lettura input: %w", err)
	}

	if err := openBrowser(authURL); err != nil {
		log.Printf("Error opening browser: %v", err)
	}

	// Block until callback handler exits the process (or server is stopped)
	select {}
}

func generatePKCE() (verifier, challenge string) {
	const cs = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	verifier = randomStringCharset(64, cs)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return
}

func randomString(n int) string {
	return randomStringCharset(n, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
}

func randomStringCharset(n int, cs string) string {
	b := make([]byte, n)
	for i := range b {
		_, _ = rand.Read(b[i : i+1])
		b[i] = cs[int(b[i])%len(cs)]
	}
	return string(b)
}

// Starts http://localhost:4000/callback and returns a stop() func to shutdown.
// Minimal timeouts and context to avoid hanging.
func startAuthCodeServer(verifier string) (func(), error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		if state != generatedState {
			http.Error(w, "Invalid state", http.StatusBadRequest)
			log.Fatalf("State mismatch: got %q", state)
		}
		if authCode == "" {
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}

		tkn := exchangeAuthCode(
			viper.GetString("token_endpoint"),
			viper.GetString(utils.DhCoreClientId),
			verifier,
			authCode,
		)
		if tkn == nil {
			http.Error(w, "Failed token exchange", http.StatusInternalServerError)
			return
		}

		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, tkn, "", "  "); err != nil {
			prettyJSON.Write(tkn)
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, "<div style=\"margin: 24px 0px 0px 24px;\">")
		fmt.Fprintln(w, "<h1>Authorization successful</h1>")
		fmt.Fprintln(w, "<h3>You may now close this window.</h3>")
		fmt.Fprintln(w, "<h3>Token response:</h3>")
		fmt.Fprintln(w, "<button style=\"position: absolute;left: 810px;padding: 10px;opacity: 0.90;cursor: pointer;\" onclick=\"navigator.clipboard.writeText(document.getElementById('resp').innerHTML)\">Copy</button>")
		fmt.Fprintf(w, "<pre id=\"resp\" style=\"background:#f6f8fa;border:1px solid #ccc;padding:16px;width:800px;min-height:400px;overflow:auto;\">%s</pre>", prettyJSON.String())
		fmt.Fprintln(w, "</div>")

		// Map token response into Viper
		var m map[string]interface{}
		if err := json.Unmarshal(tkn, &m); err != nil {
			log.Printf("json parse error: %v", err)
		}
		for k, v := range m {
			key := k
			if mapped, ok := utils.DhCoreMap[k]; ok {
				key = mapped
			}
			viper.Set(key, fmt.Sprint(v))
		}

		// I can also pass just some specific config keys
		if err := utils.UpdateIniSectionFromViper(viper.AllKeys()); err != nil {
			log.Printf("persist error: %v", err)
		}

		log.Println("Login successful.")
		go os.Exit(0)
	})

	srv := &http.Server{
		Addr:              ":4000",
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
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("auth server error: %v", err)
		}
	}()

	stop := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}

	return stop, nil
}

func exchangeAuthCode(tokenURL, clientID, verifier, code string) []byte {
	v := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"code_verifier": {verifier},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	// HTTP client with timeout
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.PostForm(tokenURL, v)
	if err != nil {
		log.Printf("Token request error: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Token error %s: %s", resp.Status, body)
		return nil
	}
	tkn, _ := io.ReadAll(resp.Body)
	return tkn
}

func buildAuthURL(chal, state string) (string, error) {
	// Robust scope normalization (commas/spaces → single space)
	raw := viper.GetString("scopes_supported")
	var scopes []string
	if raw != "" {
		split := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' })
		for _, s := range split {
			if s != "" {
				scopes = append(scopes, s)
			}
		}
	}
	scope := url.QueryEscape(strings.Join(scopes, " "))

	v := url.Values{
		"response_type":         {"code"},
		"client_id":             {viper.GetString(utils.DhCoreClientId)},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {chal},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}
	base := viper.GetString("authorization_endpoint")
	if base == "" {
		return "", fmt.Errorf("authorization_endpoint non configurato")
	}
	return base + "?" + v.Encode() + "&scope=" + scope, nil
}

func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	case "darwin":
		cmd = exec.Command("open", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	return cmd.Start()
}
