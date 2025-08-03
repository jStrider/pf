package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"pf/internal/config"
)

// NewStoreCommand creates the store command
func NewStoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "store",
		Short: "Manage password stores",
		Long:  `List, create, and manage password stores`,
	}

	cmd.AddCommand(
		newStoreListCommand(),
		newStoreAddCommand(),
		newStoreRemoveCommand(),
		newStoreSetDefaultCommand(),
	)

	return cmd
}

func newStoreListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all stores",
		RunE:  runStoreList,
	}
}

func newStoreAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [name]",
		Short: "Add a new store",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoreAdd,
	}

	cmd.Flags().String("path", "", "Store path (default: ~/.pf/stores/[name])")
	cmd.Flags().StringSlice("recipients", []string{}, "Age recipients")

	return cmd
}

func newStoreRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [name]",
		Short: "Remove a store",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoreRemove,
		ValidArgsFunction: storeNameCompletion,
	}
}

func newStoreSetDefaultCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default [name]",
		Short: "Set default store",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoreSetDefault,
		ValidArgsFunction: storeNameCompletion,
	}
}

func runStoreList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Stores) == 0 {
		cmd.Println("No stores configured")
		return nil
	}

	cmd.Println("Password stores:")
	for name, store := range cfg.Stores {
		defaultMarker := ""
		if name == cfg.DefaultStore {
			defaultMarker = " (default)"
		}
		cmd.Printf("  %s%s\n", name, defaultMarker)
		cmd.Printf("    Path: %s\n", store.Path)
		if len(store.Recipients) > 0 {
			cmd.Printf("    Recipients: %s\n", strings.Join(store.Recipients, ", "))
		}
	}

	return nil
}

func runStoreAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if store already exists
	if _, exists := cfg.Stores[name]; exists {
		return fmt.Errorf("store '%s' already exists", name)
	}

	// Get store path
	path := cmd.Flag("path").Value.String()
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(home, ".pf", "stores", name)
	}

	// Create store directory
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Get recipients
	recipients, _ := cmd.Flags().GetStringSlice("recipients")
	if len(recipients) == 0 {
		// Use default recipients from config
		if defaultStore, ok := cfg.Stores[cfg.DefaultStore]; ok {
			recipients = defaultStore.Recipients
		}
	}

	// Add store to config
	if cfg.Stores == nil {
		cfg.Stores = make(map[string]config.StoreConfig)
	}
	cfg.Stores[name] = config.StoreConfig{
		Path:       path,
		Recipients: recipients,
	}

	// Set as default if it's the first store
	if len(cfg.Stores) == 1 {
		cfg.DefaultStore = name
	}

	// Save config
	if err := saveConfig(cfg); err != nil {
		return err
	}

	// Create .recipients file
	if len(recipients) > 0 {
		recipientsFile := filepath.Join(path, ".recipients")
		recipientsData := fmt.Sprintf("# Age recipients for store '%s'\n%s\n", 
			name, strings.Join(recipients, "\n"))
		if err := os.WriteFile(recipientsFile, []byte(recipientsData), 0600); err != nil {
			return fmt.Errorf("failed to create recipients file: %w", err)
		}
	}

	cmd.Printf("Store '%s' added successfully\n", name)
	return nil
}

func runStoreRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if store exists
	if _, exists := cfg.Stores[name]; !exists {
		return fmt.Errorf("store '%s' not found", name)
	}

	// Remove from config
	delete(cfg.Stores, name)

	// Update default store if needed
	if cfg.DefaultStore == name {
		cfg.DefaultStore = ""
		// Set first remaining store as default
		for storeName := range cfg.Stores {
			cfg.DefaultStore = storeName
			break
		}
	}

	// Save config
	if err := saveConfig(cfg); err != nil {
		return err
	}

	cmd.Printf("Store '%s' removed from configuration\n", name)
	cmd.Println("Note: Store directory was not deleted")
	return nil
}

func runStoreSetDefault(cmd *cobra.Command, args []string) error {
	name := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if store exists
	if _, exists := cfg.Stores[name]; !exists {
		return fmt.Errorf("store '%s' not found", name)
	}

	// Update default store
	cfg.DefaultStore = name

	// Save config
	if err := saveConfig(cfg); err != nil {
		return err
	}

	cmd.Printf("Default store set to '%s'\n", name)
	return nil
}

func saveConfig(cfg *config.Config) error {
	configPath := cfg.GetConfigPath()
	
	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}