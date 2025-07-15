// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"dhcli/utils"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

func DeleteHandler(env string, project string, name string, confirm bool, cascade bool, resource string, id string) error {

	endpoint := utils.TranslateEndpoint(resource)

	// Load environment and check API level requirements
	utils.CheckUpdateEnvironment()
	utils.CheckApiLevel(utils.ApiLevelKey, utils.DeleteMin, utils.DeleteMax)

	if endpoint == "projects" && cascade != true {
		log.Println("WARNING: You are deleting a project without the cascade (-c) flag. Resources belonging to the project will not be deleted.")
	}

	if endpoint != "projects" && project == "" {
		log.Println("Project is mandatory when performing this operation on resources other than projects.")
		os.Exit(1)
	}

	params := map[string]string{}
	params["cascade"] = "false"
	if cascade != false {
		params["cascade"] = "true"
	}

	confirmationMessage := fmt.Sprintf("Resource %v (%v) will be deleted, proceed? Y/n", id, endpoint)
	if id == "" {
		if name == "" {
			log.Println("You must specify id or name.")
			os.Exit(1)
		}
		if endpoint != "projects" {
			confirmationMessage = fmt.Sprintf("All versions of endpoint named '%v' (%v) will be deleted, proceed? Y/n", name, endpoint)
			params["name"] = name
		} else {
			confirmationMessage = fmt.Sprintf("Resource %v (%v) will be deleted, proceed? Y/n", name, endpoint)
			id = name
		}
	}

	// Ask for confirmation
	if confirm != true {
		utils.WaitForConfirmation(confirmationMessage)
	}

	method := "DELETE"
	url := utils.BuildCoreUrl(project, endpoint, id, params)
	req := utils.PrepareRequest(method, url, nil, viper.GetString("access_token"))
	_, err := utils.DoRequest(req)
	if err != nil {
		return err
	}
	log.Println("Deleted successfully.")

	return nil
}
