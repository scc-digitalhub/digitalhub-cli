package utils

import (
	"strings"

	"github.com/google/uuid"
)

func UUIDv4NoDash() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
