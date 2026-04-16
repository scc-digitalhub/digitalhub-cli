// SPDX-FileCopyrightText: © 2025 DSLab - Fondazione Bruno Kessler
//
// SPDX-License-Identifier: Apache-2.0

package flags

import (
	"github.com/spf13/cobra"
)

type AllowedTypes interface {
	string | bool | int | float64
}

type FlagStruct[T AllowedTypes] struct {
	Name         string
	Short        string
	Long         string
	Description  string
	DefaultValue T
	Optional     bool
	Hidden       bool
	Aliases      []string
	Validation   func(T) error
	Value        *T
}

// === Factory helpers ===

func NewStringFlag(name, short, desc, def string) FlagStruct[string] {
	return FlagStruct[string]{
		Name:         name,
		Short:        short,
		Description:  desc,
		DefaultValue: def,
		Value:        new(string),
	}
}

func NewBoolFlag(name, short, desc string, def bool) FlagStruct[bool] {
	return FlagStruct[bool]{
		Name:         name,
		Short:        short,
		Description:  desc,
		DefaultValue: def,
		Value:        new(bool),
	}
}

// === We can implement more helper here ===
// ...

// === Register flag to cobra ===

func AddFlag[T AllowedTypes](cmd *cobra.Command, flag *FlagStruct[T]) {
	val := any(flag.Value)
	def := any(flag.DefaultValue)

	switch v := val.(type) {
	case *string:
		cmd.Flags().StringVarP(v, flag.Name, flag.Short, def.(string), flag.Description)
	case *bool:
		cmd.Flags().BoolVarP(v, flag.Name, flag.Short, def.(bool), flag.Description)
	case *int:
		cmd.Flags().IntVarP(v, flag.Name, flag.Short, def.(int), flag.Description)
	case *float64:
		cmd.Flags().Float64VarP(v, flag.Name, flag.Short, def.(float64), flag.Description)
	default:
		panic("unsupported flag type")
	}

	fs := cmd.Flags().Lookup(flag.Name)
	if fs != nil && flag.Hidden {
		fs.Hidden = true
	}
}
