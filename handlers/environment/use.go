// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"fmt"
	"log"
	"os"

	"dhcli/handlers/utils"
)

func UseEnv(env string) error {
	if env == "" {
		return fmt.Errorf("environment name is required")
	}
	cfg, err := utils.LoadIniErr(false)
	if err != nil {
		return err
	}
	if !cfg.HasSection(env) {
		return fmt.Errorf("specified environment %q does not exist", env)
	}

	defaultSection := cfg.Section("DEFAULT")
	defaultSection.Key("current_environment").SetValue(env)

	if err := utils.SaveIniErr(cfg); err != nil {
		return err
	}
	log.Printf("Switched default to '%v'.\n", env)
	return nil
}

func UseHandler(env string) {
	if err := UseEnv(env); err != nil {
		log.Printf("%v\n", err)
		os.Exit(1)
	}
}
