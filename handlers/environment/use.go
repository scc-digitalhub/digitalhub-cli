// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"fmt"
	"log"

	"dhcli/handlers/utils"
)

func UseHandler(env string) error {
	if env == "" {
		return fmt.Errorf("environment name is required")
	}
	cfg, err := utils.LoadIni(false)
	if err != nil {
		return err
	}
	if !cfg.HasSection(env) {
		return fmt.Errorf("specified environment %q does not exist", env)
	}

	defaultSection := cfg.Section("DEFAULT")
	defaultSection.Key("current_environment").SetValue(env)

	if err := utils.SaveIni(cfg); err != nil {
		return err
	}
	log.Printf("Switched default to '%v'.\n", env)
	return nil
}
