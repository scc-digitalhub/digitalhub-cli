// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// InitEnvironmentHandler initializes a Python virtual environment at the given path
// and installs dependencies. If the path already contains a valid venv, it just installs
// dependencies. If not, it creates a new venv first.
func InitEnvironmentHandler(venvPath string, includePre bool) error {
	// Normalize path
	absPath, err := filepath.Abs(venvPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	log.Printf("Using venv path: %s", absPath)

	// Check Python version
	out, err := exec.Command("python3", "--version").Output()
	if err != nil {
		return fmt.Errorf("python3 not found: %w", err)
	}
	if !supportedPythonVersion(string(out)) {
		return fmt.Errorf("unsupported Python version (need 3.9–3.12): %s", strings.TrimSpace(string(out)))
	}

	// Check if venv already exists and is valid
	pythonExe := filepath.Join(absPath, "bin", "python")
	venvExists := isValidVenv(absPath)

	if venvExists {
		log.Printf("Found existing venv at %s", absPath)
	} else {
		log.Printf("Creating new venv at %s", absPath)
		if err := createVenv(absPath); err != nil {
			return fmt.Errorf("failed to create venv: %w", err)
		}
	}

	// Get API version
	apiVer := viper.GetString("dhcore_version")
	parts := strings.SplitN(apiVer, ".", 3)
	if len(parts) > 2 {
		apiVer = parts[0] + "." + parts[1]
	}

	// Confirmation prompt
	yes := promptYesNo(fmt.Sprintf("Newest patch version of digitalhub %v will be installed, continue? Y/n", apiVer))
	if !yes {
		log.Println("Installation cancelled by user.")
		return nil
	}

	// Build pip command
	// If --pre is passed, allow beta versions. Otherwise, require stable versions.
	nextMinor := incrementMinorVersion(apiVer)
	var pipSpec string
	if includePre {
		pipSpec = ">=" + apiVer + ".0b0,<" + nextMinor
	} else {
		pipSpec = ">=" + apiVer + ".0,<" + nextMinor
	}

	// Install packages using the venv's python/pip
	for _, pkg := range packageList() {
		args := []string{"-m", "pip", "install", pkg + pipSpec}
		if includePre {
			args = append(args, "--pre")
		}

		cmd := exec.Command(pythonExe, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Printf("Installing %s (in venv %s)...", pkg+pipSpec, absPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pip install failed for %s: %w", pkg, err)
		}
	}

	log.Println("Installation complete.")
	log.Printf("To activate the venv, run: source %s/bin/activate", absPath)
	return nil
}

// isValidVenv checks if a directory contains a valid Python virtual environment
func isValidVenv(path string) bool {
	pythonExe := filepath.Join(path, "bin", "python")
	pyvenvCfg := filepath.Join(path, "pyvenv.cfg")

	// Check if python executable exists and pyvenv.cfg exists
	_, err1 := os.Stat(pythonExe)
	_, err2 := os.Stat(pyvenvCfg)

	return err1 == nil && err2 == nil
}

// createVenv creates a new Python virtual environment at the given path
func createVenv(path string) error {
	// Check if path already exists
	if info, err := os.Stat(path); err == nil {
		// Path exists - check if it's a valid venv
		if isValidVenv(path) {
			// Valid venv already exists, nothing to do
			return nil
		}
		// Path exists but is not a valid venv - this is an error
		if info.IsDir() {
			log.Printf("⚠  ERROR: Path %s already exists but is not a valid Python venv", path)
			log.Printf("Please remove or choose a different path.")
			return fmt.Errorf("path exists but is not a valid venv: %s", path)
		}
		// Path exists as a file, not a directory
		log.Printf("⚠  ERROR: Path %s exists but is a file, not a directory", path)
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Path doesn't exist - create the venv
	cmd := exec.Command("python3", "-m", "venv", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("venv creation failed: %w", err)
	}

	return nil
}

func supportedPythonVersion(ver string) bool {
	ver = strings.TrimSpace(ver)
	if idx := strings.Index(ver, " "); idx >= 0 && len(ver) > idx+1 {
		ver = ver[idx+1:]
	}
	parts := strings.Split(ver, ".")
	if len(parts) < 2 {
		return false
	}
	maj, err := strconv.Atoi(parts[0])
	if err != nil || maj != 3 {
		return false
	}
	min, err := strconv.Atoi(parts[1])
	if err != nil || min < 9 || min > 12 {
		return false
	}
	return true
}

func promptYesNo(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(prompt)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "" {
			return true
		}
		if input == "n" {
			return false
		}
		fmt.Print("Invalid input, please type Y or n: ")
	}
}

func packageList() []string {
	return []string{"digitalhub[full]", "digitalhub-runtime-python"}
}

// incrementMinorVersion increments the minor version number
// e.g., "0.15" -> "0.16"
func incrementMinorVersion(apiVer string) string {
	parts := strings.Split(apiVer, ".")
	if len(parts) < 2 {
		return apiVer // Fallback if format unexpected
	}
	major := parts[0]
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return apiVer // Fallback if parse fails
	}
	return major + "." + strconv.Itoa(minor+1)
}
