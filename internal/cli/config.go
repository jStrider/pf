package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"pf/internal/config"
)

// NewConfigCommand creates the config command
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  `View and edit configuration settings`,
	}

	cmd.AddCommand(
		newConfigShowCommand(),
		newConfigGetCommand(),
		newConfigSetCommand(),
	)

	return cmd
}

func newConfigShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show all configuration",
		RunE:  runConfigShow,
	}
}

func newConfigGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE:  runConfigGet,
	}
}

func newConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE:  runConfigSet,
	}
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Marshal to YAML for pretty printing
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	cmd.Printf("Configuration:\n%s", string(data))
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "default_store":
		cmd.Println(cfg.DefaultStore)
	case "age_key_path":
		cmd.Println(cfg.AgeKeyPath)
	case "audit_log":
		cmd.Println(cfg.AuditLog)
	case "clipboard_timeout":
		cmd.Printf("%s\n", cfg.ClipboardTimeout)
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch key {
	case "default_store":
		// Validate store exists
		if _, ok := cfg.Stores[value]; !ok {
			return fmt.Errorf("store '%s' not found", value)
		}
		cfg.DefaultStore = value
	case "age_key_path":
		cfg.AgeKeyPath = value
	case "audit_log":
		cfg.AuditLog = value == "true"
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// Save config
	if err := saveConfig(cfg); err != nil {
		return err
	}

	cmd.Printf("Configuration updated: %s = %s\n", key, value)
	return nil
}