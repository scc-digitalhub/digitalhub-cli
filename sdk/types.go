// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package sdk

// Richiesta/risposta per download (usata dall’SDK e dall’adapter)
type DownloadRequest struct {
	Project     string
	Resource    string // canonical (es: "artifacts")
	ID          string // opzionale
	Name        string // usato se ID vuoto
	Destination string // file o directory
	Verbose     bool   // abilita progress/hook nel download S3
}

type DownloadInfo struct {
	Filename string `json:"filename" yaml:"filename"`
	Size     int64  `json:"size"     yaml:"size"`
	Path     string `json:"path"     yaml:"path"`
}
