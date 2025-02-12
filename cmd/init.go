package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
)

func init() {
	RegisterCommand(&Command{
		Name:        "init",
		Description: "./dhcli init <environment>",
		SetupFlags:  func(fs *flag.FlagSet) {},
		Handler:     initHandler,
	})
}

func initHandler(args []string, fs *flag.FlagSet) {
	ini.DefaultHeader = true

	// Check if Python version is supported
	versionOutput, err := exec.Command("bash", "-c", "python --version").Output()
	if err != nil {
		log.Fatalf("Failed to retrieve Python version: %v", err)
	}
	if !supportedPythonVersion(string(versionOutput)) {
		log.Fatalf("Python version is not supported (must be 3.9.xx <= v <=3.12.xx): %v", string(versionOutput))
	}

	// Check if pip is installed
	_, err = exec.Command("bash", "-c", "pip --version").Output()
	if err != nil {
		log.Fatalf("Failed to retrieve pip version: %v", err)
	}

	// Read config from ini file
	_, section := loadConfig(args)

	apiVersion := section.Key("version").String()
	apiVersionMinor := apiVersion[:strings.LastIndex(apiVersion, ".")]

	// Ask for confirmation
	for {
		buf := bufio.NewReader(os.Stdin)
		fmt.Printf("Newest patch version of digitalhub %v will be installed, continue? Y/n\n", apiVersionMinor)
		userInput, err := buf.ReadBytes('\n')
		if err != nil {
			log.Fatalf("Error in reading user input: %v", err)
		} else {
			yn := strings.TrimSpace(string(userInput))
			if strings.ToLower(yn) == "y" || yn == "" {
				break
			} else if strings.ToLower(yn) == "n" {
				fmt.Println("Cancelling installation.")
				return
			}
			fmt.Println("Invalid input, must be y or n")
		}
	}

	cmd := "pip install digitalhub~=" + apiVersionMinor
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Fatalf("Failed to execute command: %v; %v", err, string(out[:]))
	}
	fmt.Println(string(out))
}

func supportedPythonVersion(pythonVersion string) bool {
	// Remove 'Python ' part
	ver := pythonVersion[strings.Index(pythonVersion, " ")+1:]

	// Split and check major and minor
	majorString := ver[:strings.Index(ver, ".")]
	if majorString != "3" {
		return false
	}

	minorString := ver[strings.Index(ver, ".")+1 : strings.LastIndex(ver, ".")]
	minor, err := strconv.Atoi(minorString)
	if err != nil || minor < 9 || minor > 12 {
		return false
	}

	return true
}
