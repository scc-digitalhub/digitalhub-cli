// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"dhcli/pkg"
	_ "dhcli/pkg/cmd"
	"log"

	"gopkg.in/ini.v1"
)

func main() {

	log.SetFlags(0)
	ini.DefaultHeader = true
	core.Execute()
}
