// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"

	"gopkg.in/ini.v1"
)

func getIniPath() string {
	iniPath, err := os.UserHomeDir()
	if err != nil {
		iniPath = "."
	}
	iniPath += string(os.PathSeparator) + IniName

	return iniPath
}

func LoadIni(createOnMissing bool) *ini.File {
	cfg, err := ini.Load(getIniPath())
	if err != nil {
		if !createOnMissing {
			log.Printf("Failed to read ini file: %v\n", err)
			os.Exit(1)
		}
		return ini.Empty()
	}

	return cfg
}

func SaveIni(cfg *ini.File) {
	err := cfg.SaveTo(getIniPath())
	if err != nil {
		log.Printf("Failed to update ini file: %v\n", err)
		os.Exit(1)
	}
}

func ReflectValue(v interface{}) string {
	f := reflect.ValueOf(v)

	switch f.Kind() {
	case reflect.String:
		return f.String()
	case reflect.Int, reflect.Int64:
		return fmt.Sprint(f.Int())
	case reflect.Uint, reflect.Uint64:
		return fmt.Sprint(f.Uint())
	case reflect.Float64:
		return fmt.Sprint(f.Float())
	case reflect.Bool:
		return fmt.Sprint(f.Bool())
	case reflect.TypeOf(time.Now()).Kind():
		return f.Interface().(time.Time).Format(time.RFC3339)
	case reflect.Slice:
		s := []string{}
		for _, element := range f.Interface().([]interface{}) {
			if reflect.ValueOf(element).Kind() == reflect.String {
				s = append(s, element.(string))
			}
		}
		return strings.Join(s, ",")
	default:
		return ""
	}
}

func BuildCoreUrl(project string, resource string, id string, params map[string]string) string {
	base := viper.GetString(DhCoreEndpoint) + "/api/" + viper.GetString(DhCoreApiVersion)
	endpoint := ""
	paramsString := ""
	if resource != "projects" && project != "" {
		endpoint += "/-/" + project
	}
	endpoint += "/" + resource
	if id != "" {
		endpoint += "/" + id
	}
	if params != nil && len(params) > 0 {
		paramsString = "?"
		for key, val := range params {
			if val != "" {
				paramsString += key + "=" + val + "&"
			}
		}
		paramsString = paramsString[:len(paramsString)-1]
	}

	return base + endpoint + paramsString
}

func PrepareRequest(method string, url string, data []byte, accessToken string) *http.Request {
	var body io.Reader = nil
	if data != nil {
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("Failed to initialize request: %v\n", err)
		os.Exit(1)
	}

	if data != nil {
		req.Header.Add("Content-type", "application/json")
	}

	if accessToken != "" {
		req.Header.Add("Authorization", "Bearer "+accessToken)
	}

	return req
}

func DoRequest(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error performing request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		// Extract message from body (if present), to return a more meaningful error
		msg := ""
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			if message, ok := bodyMap["message"]; ok && reflect.ValueOf(message).Kind() == reflect.String {
				msg += " - " + message.(string)
			}
		}

		log.Printf("Core responded with: %v%v\n", resp.Status, msg)
		os.Exit(1)
	}

	return body, err
}

func TranslateFormat(format string) string {
	lower := strings.ToLower(format)
	if lower == "json" {
		return "json"
	} else if lower == "yaml" || lower == "yml" {
		return "yaml"
	}
	return "short"
}

func LoadIniConfig(args []string) (*ini.File, *ini.Section) {
	cfg := LoadIni(false)

	sectionName := ""

	if len(args) == 0 || args[0] == "" {
		if cfg.HasSection("DEFAULT") {
			defaultSection, err := cfg.GetSection("DEFAULT")
			if err != nil {
				log.Printf("Error while reading default environment: %v\n", err)
				os.Exit(1)
			}
			if defaultSection.HasKey("current_environment") {
				sectionName = defaultSection.Key("current_environment").String()
			}
		}

		if sectionName == "" {
			log.Println("Error: environment was not passed and default environment is not specified in ini file.")
			os.Exit(1)
		}
	} else {
		sectionName = args[0]
	}

	section, err := cfg.GetSection(sectionName)
	if err != nil {
		log.Printf("Failed to read section '%s': %v.\n", sectionName, err)
		os.Exit(1)
	}

	return cfg, section
}

func TranslateEndpoint(resource string) string {
	for key, val := range Resources {
		if key == resource || slices.Contains(val, resource) {
			return key
		}
	}

	log.Printf("Resource '%v' is not supported.\n", resource)
	os.Exit(1)
	return ""
}

func GetFirstIfList(m map[string]interface{}) (map[string]interface{}, error) {
	if content, ok := m["content"]; ok && reflect.ValueOf(content).Kind() == reflect.Slice {
		contentSlice := content.([]interface{})
		if len(contentSlice) >= 1 {
			return contentSlice[0].(map[string]interface{}), nil
		}
		return nil, errors.New("Resource not found")
	}
	return m, nil
}

func WaitForConfirmation(msg string) {
	for {
		buf := bufio.NewReader(os.Stdin)
		log.Printf(msg)
		userInput, err := buf.ReadBytes('\n')
		if err != nil {
			log.Printf("Error in reading user input: %v\n", err)
			os.Exit(1)
		} else {
			yn := strings.TrimSpace(string(userInput))
			if strings.ToLower(yn) == "y" || yn == "" {
				break
			} else if strings.ToLower(yn) == "n" {
				log.Println("Cancelling.")
				os.Exit(0)
			}
			log.Println("Invalid input, must be y or n")
		}
	}
}

func PrintCommentForYaml(args ...string) {
	fmt.Printf("# Generated on: %v\n", time.Now().Round(0))
	fmt.Printf("#   from environment: %v (core version %v)\n", viper.GetString("dhcore_name"), viper.GetString("dhcore_version"))
	fmt.Printf("#   found at: %v\n", viper.GetString(DhCoreEndpoint))
	argsString := ""
	for _, s := range args {
		if s != "" {
			argsString += s + " "
		}
	}
	if argsString != "" {
		fmt.Printf("#   with parameters: %v\n", argsString[:len(argsString)-1])
	}
}

func CheckApiLevel(apiLevelKey string, min int, max int) {

	//print api level key
	fmt.Printf("Checking API level for %v command...\n", viper.GetString(apiLevelKey))

	apiLevelKeyString := viper.GetString(apiLevelKey)

	if apiLevelKeyString == "" {
		log.Println("ERROR: Unable to check compatibility, environment does not specify API level.")
		os.Exit(1)
	}

	apiLevel, err := strconv.Atoi(apiLevelKeyString)
	if err != nil {
		log.Printf("ERROR: Unable to check compatibility, as API level %v could not be read as integer.\n", apiLevelKeyString)
		os.Exit(1)
	}

	supportedInterval := ""
	if min != 0 {
		supportedInterval += fmt.Sprintf("%v <= ", min)
	}
	supportedInterval += "level"
	if max != 0 {
		supportedInterval += fmt.Sprintf(" <= %v", max)
	}

	if (min != 0 && apiLevel < min) || (max != 0 && apiLevel > max) {
		log.Printf("ERROR: API level %v is not within the supported interval for this command: %v\n", apiLevel, supportedInterval)
		os.Exit(1)
	}
}

func GetStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

func FetchConfig(configURL string) (map[string]interface{}, error) {
	resp, err := http.Get(configURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Core returned a non-200 status code: %v", resp.Status))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func PrintResponseState(resp []byte) error {
	// Parse response to check new state
	var m map[string]interface{}
	if err := json.Unmarshal(resp, &m); err != nil {
		return err
	}
	if status, ok := m["status"]; ok {
		statusMap := status.(map[string]interface{})
		if state, ok := statusMap["state"]; ok {
			log.Printf("Core response successful, new state: %v\n", state.(string))
			return nil
		}
	}

	log.Println("WARNING: core response successful, but unable to confirm new state.")
	return nil
}
