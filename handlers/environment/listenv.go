// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"log"

	"dhcli/handlers/utils"
)

func ListEnvHandler() {
	cfg := utils.LoadIni(true)

	currentEnv := cfg.Section("DEFAULT").Key("current_environment").String()
	if currentEnv != "" {
		log.Printf("Current environment: %s\n", currentEnv)
	}

	sections := cfg.SectionStrings()
	sectionsString := ""

	for _, name := range sections {
		if name != "DEFAULT" {
			sectionsString += name + ", "
		}
	}

	if sectionsString == "" {
		log.Println("No environments available.")
		return
	}
	sectionsString = sectionsString[:len(sectionsString)-2]

	log.Printf("Available environments: %s\n", sectionsString)
}
