// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package facade

import (
	"dhcli/utils"
	"log"
	"os"
)

func UseHandler(env string) {
	environmentName := env
	cfg := utils.LoadIni(false)
	if !cfg.HasSection(environmentName) {
		log.Printf("Specified environment does not exist.\n")
		os.Exit(1)
	}

	defaultSection := cfg.Section("DEFAULT")
	defaultSection.Key("current_environment").SetValue(environmentName)

	utils.SaveIni(cfg)
	log.Printf("Switched default to '%v'.\n", environmentName)
}
