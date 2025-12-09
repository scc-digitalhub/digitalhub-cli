// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package facade

import (
	"log"
	"os"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/utils"
)

func RemoveHandler(env string) {
	sectionName := env

	cfg := utils.LoadIni(false)
	if !cfg.HasSection(sectionName) {
		log.Printf("Specified environment does not exist.\n")
		os.Exit(0)
	}

	cfg.DeleteSection(sectionName)

	defaultSection := cfg.Section("DEFAULT")
	if defaultSection.Key("current_environment").String() == sectionName {
		defaultSection.DeleteKey("current_environment")
	}

	utils.SaveIni(cfg)
	log.Printf("'%v' has been removed.\n", sectionName)
}
