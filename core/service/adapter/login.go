// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package adapter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/viper"

	"dhcli/sdk"
	"dhcli/utils"
)

const redirectURI = "http://localhost:4000/callback"

var generatedState string

// Runs PKCE flow for authentication (stessa UX e stesso HTML di prima)
func LoginHandler() error {
	// Ensure environment is up-to-date and compatible
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.LoginMin, utils.LoginMax)

	// PKCE (SDK)
	pkce := sdk.NewPKCE()
	generatedState = pkce.State

	// Server locale (SDK) con callback che:
	// - stampa HTML con token pretty
	// - mappa i token in viper (DhCoreMap)
	// - persiste INI
	// - logga "Login successful." e fa os.Exit(0)
	stop, err := sdk.StartAuthCodeServer(
		sdk.AuthConfig{
			AuthorizationEndpoint: viper.GetString(utils.Oauth2AuthorizationEndpoint),
			TokenEndpoint:         viper.GetString(utils.Oauth2TokenEndpoint),
			ClientID:              viper.GetString(utils.DhCoreClientId),
			RedirectURI:           redirectURI,
			Scopes:                normalizeScopes(viper.GetString("scopes_supported")), // usato per URL
		},
		pkce,
		func(tokens []byte, w http.ResponseWriter) {
			// === HTML identico all’originale ===
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, tokens, "", "  "); err != nil {
				prettyJSON.Write(tokens)
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintln(w, "<div style=\"margin: 24px 0px 0px 24px;\">")
			fmt.Fprintln(w, "<h1>Authorization successful</h1>")
			fmt.Fprintln(w, "<h3>You may now close this window.</h3>")
			fmt.Fprintln(w, "<h3>Token response:</h3>")
			fmt.Fprintln(w, "<button style=\"position: absolute;left: 810px;padding: 10px;opacity: 0.90;cursor: pointer;\" onclick=\"navigator.clipboard.writeText(document.getElementById('resp').innerHTML)\">Copy</button>")
			fmt.Fprintf(w, "<pre id=\"resp\" style=\"background:#f6f8fa;border:1px solid #ccc;padding:16px;width:800px;min-height:400px;overflow:auto;\">%s</pre>", prettyJSON.String())
			fmt.Fprintln(w, "</div>")

			// === Mapping token -> viper (come prima) ===
			var m map[string]interface{}
			if err := json.Unmarshal(tokens, &m); err != nil {
				log.Printf("json parse error: %v", err)
			}
			for k, v := range m {
				key := k
				if mapped, ok := utils.DhCoreMap[k]; ok {
					key = mapped
				}
				viper.Set(key, utils.ReflectValue(v))
			}

			// Aggiorno freshness e persisto INI (identico)
			viper.Set(utils.UpdatedEnvKey, time.Now().UTC().Format(time.RFC3339))
			if err := utils.UpdateIniSectionFromViper(viper.AllKeys()); err != nil {
				log.Printf("persist error: %v", err)
			}

			log.Println("Login successful.")
			go os.Exit(0)
		},
		func(err error, w http.ResponseWriter) {
			// Errore come prima (testuale)
			http.Error(w, "Failed token exchange", http.StatusInternalServerError)
			log.Printf("Login error: %v", err)
		},
	)
	if err != nil {
		return fmt.Errorf("impossibile avviare il server locale: %w", err)
	}
	defer stop()

	// Build authorize URL (identico: stessi scope, stessi parametri)
	authURL, err := buildAuthURLViaSDK(
		sdk.AuthConfig{
			AuthorizationEndpoint: viper.GetString(utils.Oauth2AuthorizationEndpoint),
			TokenEndpoint:         viper.GetString(utils.Oauth2TokenEndpoint),
			ClientID:              viper.GetString(utils.DhCoreClientId),
			RedirectURI:           redirectURI,
			Scopes:                normalizeScopes(viper.GetString("scopes_supported")),
		},
		pkce,
	)
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

	// Block until callback handler exits the process (identico)
	select {}
}

// Build URL con le stesse regole dell’originale (scope normalizzato e parametri)
func buildAuthURLViaSDK(cfg sdk.AuthConfig, pkce *sdk.PKCE) (string, error) {
	// Normalizzazione scope già fatta in normalizeScopes(); qui solo encoding.
	scope := url.QueryEscape(strings.Join(cfg.Scopes, " "))
	v := url.Values{
		"response_type":         {"code"},
		"client_id":             {cfg.ClientID},
		"redirect_uri":          {cfg.RedirectURI},
		"code_challenge":        {pkce.Challenge},
		"code_challenge_method": {"S256"},
		"state":                 {pkce.State},
	}
	base := cfg.AuthorizationEndpoint
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

func normalizeScopes(raw string) []string {
	if raw == "" {
		return nil
	}
	f := func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' }
	parts := strings.FieldsFunc(raw, f)
	var scopes []string
	for _, p := range parts {
		if p != "" {
			scopes = append(scopes, p)
		}
	}
	return scopes
}
