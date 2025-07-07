// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package models

type Entity struct {
	Spec Spec `json:"spec"`
}

func (a Entity) GetSpec() Spec {
	return a.Spec
}
