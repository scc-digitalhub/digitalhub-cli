// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"fmt"
	"log"

	"dhcli/handlers/utils"
)

func RemoveHandler(env string) error {
	sectionName := env

	cfg, err := utils.LoadIni(false)
	if err != nil {
		return err
	}
	if !cfg.HasSection(sectionName) {
		return fmt.Errorf("specified environment %q does not exist", sectionName)
	}

	cfg.DeleteSection(sectionName)

	defaultSection := cfg.Section("DEFAULT")
	if defaultSection.Key("current_environment").String() == sectionName {
		defaultSection.DeleteKey("current_environment")
	}

	if err := utils.SaveIni(cfg); err != nil {
		return err
	}
	log.Printf("'%v' has been removed.\n", sectionName)
	return nil
}
