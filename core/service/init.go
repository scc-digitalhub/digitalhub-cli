// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"dhcli/utils"
)

// InitEnvironmentHandler verifica Python, conferma l'utente e installa i pacchetti
func InitEnvironmentHandler(env string, includePre bool) error {
	// 1. Controllo versione Python
	out, err := exec.Command("python3", "--version").Output()
	if err != nil {
		return fmt.Errorf("python3 non trovato: %w", err)
	}
	if !supportedPythonVersion(string(out)) {
		return fmt.Errorf("versione Python non supportata (serve 3.9–3.12): %s", strings.TrimSpace(string(out)))
	}

	// 2. Legge la configurazione dall’ini
	_, section := utils.LoadIniConfig([]string{env})
	// La tua funzione mantiene lo stesso comportamento: richiama LoadIniConfig

	// 3. Estrae la minor version
	apiVer := section.Key("dhcore_version").String()
	parts := strings.SplitN(apiVer, ".", 3)
	if len(parts) > 2 {
		apiVer = parts[0] + "." + parts[1]
	}

	// 4. Prompt di conferma
	yes := promptYesNo(fmt.Sprintf("Newest patch version of digitalhub %v will be installed, continue? Y/n", apiVer))
	if !yes {
		log.Println("Installation cancelled by user.")
		return nil
	}

	// 5. Costruisce l’opzione pip
	pipSpec := "~=" + apiVer + ".0"

	// 6. Installa ogni pacchetto, nel formato originale
	for _, pkg := range packageList() {
		args := []string{"-m", "pip", "install", pkg + pipSpec}
		if includePre {
			args = append(args, "--pre")
		}

		cmd := exec.Command("python3", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		log.Printf("Installing %s...", pkg+pipSpec)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("pip install failed for %s: %w", pkg, err)
		}
	}

	log.Println("Installation complete.")
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
