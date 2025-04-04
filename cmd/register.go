package cmd

import (
	"bufio"
	"dhcli/utils"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"

	"gopkg.in/ini.v1"
)

type OpenIDConfig struct {
	AuthorizationEndpoint string   `json:"authorization_endpoint" ini:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint" ini:"token_endpoint"`
	Issuer                string   `json:"issuer" ini:"issuer"`
	ClientID              string   `json:"dhcore_client_id" ini:"dhcore_client_id"`
	Scope                 []string `json:"scopes_supported" ini:"scopes_supported"`
	AccessToken           string   `json:"access_token" ini:"access_token"`
	RefreshToken          string   `json:"refresh_token" ini:"refresh_token"`
}

type CoreConfig struct {
	Name     string `json:"dhcore_name" ini:"dhcore_name"`
	Issuer   string `json:"issuer" ini:"issuer"`
	Version  string `json:"dhcore_version" ini:"dhcore_version"`
	ClientID string `json:"dhcore_client_id" ini:"dhcore_client_id"`
}

func init() {
	RegisterCommand(&Command{
		Name:        "register",
		Description: "dhcli register [-n <name>] <endpoint>",
		SetupFlags: func(fs *flag.FlagSet) {
			fs.String("n", "", "name")
		},
		Handler: registerHandler,
	})
}

func registerHandler(args []string, fs *flag.FlagSet) {
	ini.DefaultHeader = true

	if len(args) < 1 {
		fmt.Printf("Error: Endpoint is required.\nUsage: dhcli register [-n <name>] <endpoint>\n")
		os.Exit(1)
	}
	fs.Parse(args)

	name := fs.Lookup("n").Value.String()
	endpoint := fs.Args()[0]
	if !strings.HasSuffix(endpoint, "/") {
		endpoint += "/"
	}

	// Read or initialize ini file
	cfg := utils.LoadIni(true)

	//collect to map+struct
	res, coreConfig := fetchConfig(endpoint + ".well-known/configuration")
	if name == "" || name == "null" {
		name = coreConfig.Name
		if name == "" {
			fmt.Printf("Failed to register: environment name not specified and not defined in core's configuration.\n")
			os.Exit(1)
		}
	}
	sec := cfg.Section(name)
	sec.ReflectFrom(&coreConfig)

	// Fetch OpenID configuration
	openIDConfig := fetchOpenIDConfig(endpoint + ".well-known/openid-configuration")
	openIDConfig.ClientID = coreConfig.ClientID
	sec.ReflectFrom(&openIDConfig)

	for k, v := range res {
		//add missing keys
		if !sec.HasKey(k) {
			sec.NewKey(k, utils.ReflectValue(v))
		}
	}

	//check for default env
	dsec := cfg.Section("DEFAULT")
	if !dsec.HasKey(utils.CurrentEnvironment) {
		dsec.NewKey(utils.CurrentEnvironment, name)
	}

	// gitignoreAddIniFile()
	utils.SaveIni(cfg)
	fmt.Printf("'%v' registered.\n", name)
}

func fetchConfig(configURL string) (map[string]interface{}, CoreConfig) {
	resp, err := http.Get(configURL)
	if err != nil {
		fmt.Printf("Error fetching core configuration: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Core responded with error %v\n", resp.Status)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading core configuration response: %v\n", err)
		os.Exit(1)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Printf("Error parsing core configuration: %v\n", err)
		os.Exit(1)
	}

	var config CoreConfig
	if err := json.Unmarshal(body, &config); err != nil {
		fmt.Printf("Error parsing core configuration: %v\n", err)
		os.Exit(1)
	}

	return res, config
}

func fetchOpenIDConfig(configURL string) OpenIDConfig {
	resp, err := http.Get(configURL)
	if err != nil {
		fmt.Printf("Error fetching OpenID configuration: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Core responded with error %v\n", resp.Status)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading OpenID configuration response: %v\n", err)
		os.Exit(1)
	}

	var config OpenIDConfig
	if err := json.Unmarshal(body, &config); err != nil {
		fmt.Printf("Error parsing OpenID configuration: %v\n", err)
		os.Exit(1)
	}

	return config
}

func toMap(strc interface{}) (map[string]interface{}, error) {

	res := make(map[string]interface{})

	// get or dereference
	val := reflect.ValueOf(strc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	if val.Kind() != reflect.Struct {
		return res, errors.New("variable given is not a struct or a pointer to a struct")
	}

	//export to value
	//NOTE: doesn't support nested structs
	for i := 0; i < val.NumField(); i++ {
		fName := typ.Field(i).Name
		fValue := val.Field(i).Interface()
		res[fName] = fValue
	}

	return res, nil
}

func gitignoreAddIniFile() {
	path := "./.gitignore"
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Cannot open .gitignore file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() == utils.IniName {
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error while reading .gitignore file contents: %v\n", err)
		os.Exit(1)
	}

	if _, err = f.WriteString(utils.IniName); err != nil {
		fmt.Printf("Error while adding entry to .gitignore file: %v\n", err)
		os.Exit(1)
	}
}
