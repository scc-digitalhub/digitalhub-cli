package utils

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

//// Recursively bind all flags of a command and its subcommands to Viper.
//func BindFlagsToViperRecursive(cmd *cobra.Command) {
//	// Bind local flags
//	cmd.Flags().VisitAll(func(f *pflag.Flag) {
//		_ = viper.BindPFlag(f.Name, f)
//	})
//	// Bind persistent flags
//	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
//		_ = viper.BindPFlag(f.Name, f)
//	})
//	// Recurse into subcommands
//	for _, sub := range cmd.Commands() {
//		BindFlagsToViperRecursive(sub)
//	}
//}

func setupEnv() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func RegisterIniCfgWithViper(optionalEnv ...string) error {
	iniPath := getIniPath()

	cfg, err := ini.Load(iniPath)
	if err != nil {
		// INI missing: ENV-only mode.
		fmt.Printf("⚠️  ini file not found or unreadable (%v); falling back to environment variables only\n", err)
		setupEnv()
		dumpEnvWithPrefix("") // optional: debug raw env
		return nil
	}

	defaultSection := cfg.Section("DEFAULT")
	var env string
	if len(optionalEnv) > 0 && optionalEnv[0] != "" {
		env = optionalEnv[0]
	} else {
		env = defaultSection.Key("current_environment").String()
	}

	var selected *ini.Section
	if env != "" && cfg.HasSection(env) {
		selected = cfg.Section(env)
		fmt.Printf("✅ Using section: [%s]\n", env)
	} else {
		selected = defaultSection
		if env == "" || env == "DEFAULT" {
			fmt.Println("ℹ️  no environment selected, using [DEFAULT]")
		} else {
			fmt.Println("⚠️  current_environment not found/invalid, falling back to [DEFAULT]")
		}
	}

	merged := make(map[string]string)

	for _, k := range defaultSection.Keys() {
		merged[k.Name()] = k.Value()
	}

	if selected != nil && selected != defaultSection {
		for _, k := range selected.Keys() {
			merged[k.Name()] = k.Value()
		}
	}

	buf := &bytes.Buffer{}
	for k, v := range merged {
		// minimal TOML escaping for strings
		val := strings.ReplaceAll(v, `\`, `\\`)
		val = strings.ReplaceAll(val, `"`, `\"`)
		if _, err := fmt.Fprintf(buf, "%s = \"%s\"\n", k, val); err != nil {
			return fmt.Errorf("failed to serialize ini section: %w", err)
		}
	}

	viper.SetConfigType("toml")
	if err := viper.ReadConfig(buf); err != nil {
		return fmt.Errorf("failed to load section into viper: %w", err)
	}

	viper.Set("current_environment", env)
	setupEnv()

	return nil
}

// Utility function to dump environment variables with a given prefix
func dumpEnvWithPrefix(prefix string) {
	fmt.Printf("\nEnvironment variables with prefix '%s':\n", prefix)
	found := false
	for _, e := range os.Environ() {
		if prefix == "" || strings.HasPrefix(e, prefix) {
			fmt.Println(e)
			found = true
		}
	}
	if !found {
		fmt.Println("(none found)")
	}
}
