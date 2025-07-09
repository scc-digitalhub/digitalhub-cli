// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"dhcli/core"
	_ "dhcli/core/commands"
	"gopkg.in/ini.v1"
	"log"
)

func main() {

	log.SetFlags(0)
	ini.DefaultHeader = true
	core.Execute()
}
