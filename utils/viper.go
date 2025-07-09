package utils

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
	"os"
	"strings"
)

// Recurisvely bind all flags of a command and its subcommands to Viper.
func BindFlagsToViperRecursive(cmd *cobra.Command) {

	// Bind local flags
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = viper.BindPFlag(f.Name, f)
	})

	// Bind persistent flags
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		_ = viper.BindPFlag(f.Name, f)
	})

	// Recursively bind for subcommands
	for _, sub := range cmd.Commands() {
		BindFlagsToViperRecursive(sub)
	}
}

func RegisterIniCfgWithViper(optionalEnv ...string) {
	iniPath := os.ExpandEnv("$HOME/" + IniName)
	cfg, err := ini.Load(iniPath)
	if err != nil {
		panic(fmt.Errorf("failed to read ini file: %w", err))
	}

	defaultSection := cfg.Section("DEFAULT")

	var env string
	if len(optionalEnv) > 0 && optionalEnv[0] != "" {
		env = optionalEnv[0]
	} else {
		env = defaultSection.Key("current_environment").String()
	}

	var selectedSection *ini.Section
	if env != "" && cfg.HasSection(env) {
		selectedSection = cfg.Section(env)
		fmt.Printf("✅ Using section: [%s]\n", env)
	} else {
		selectedSection = defaultSection
		fmt.Println("⚠️  current_environment not found or invalid, falling back to [DEFAULT]")
	}

	// Flatten selected section into TOML
	buf := &bytes.Buffer{}
	for _, key := range selectedSection.Keys() {
		_, err := fmt.Fprintf(buf, "%s = \"%s\"\n", key.Name(), key.Value())
		if err != nil {
			return
		}
	}

	viper.SetConfigType("toml")
	if err := viper.ReadConfig(buf); err != nil {
		panic(fmt.Errorf("failed to load section into viper: %w", err))
	}

	//set current_environment in viper
	viper.Set("current_environment", env)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}
