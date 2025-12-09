// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package facade

import (
	"fmt"
	"log"

	"github.com/scc-digitalhub/digitalhub-cli-sdk/sdk/utils"
)

func ListEnvHandler() {
	cfg := utils.LoadIni(true)
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

	log.Println("Available environments:")
	fmt.Printf("%v\n", sectionsString)
}
