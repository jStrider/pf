package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"pf/internal/age"
	"pf/internal/config"
)

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new password store",
		Long:  `Initialize a new password store with age encryption`,
		RunE:  runInit,
	}

	cmd.Flags().String("store", "", "Store name (default: personal)")
	cmd.Flags().String("age-key", "", "Path to age private key file")
	cmd.Flags().String("recipient", "", "Age recipient public key")

	return cmd
}

func runInit(cmd *cobra.Command, args []string) error {
	// Get config directory
	configDir := viper.GetString("config_dir")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configDir = filepath.Join(home, ".pf")
	}

	// Create config directory
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Get or generate age key
	var keyPair *age.Key
	ageKeyPath := cmd.Flag("age-key").Value.String()
	recipient := cmd.Flag("recipient").Value.String()

	if ageKeyPath != "" {
		// Load existing key
		identities, err := age.LoadIdentityFile(ageKeyPath)
		if err != nil {
			return fmt.Errorf("failed to load age key: %w", err)
		}
		if len(identities) == 0 {
			return fmt.Errorf("no identities found in key file")
		}
		// Use the first identity
		// Note: we can't access the string representation directly from the interface
		// We'll need to use the identity as-is
		keyPair = &age.Key{
			Identity: "loaded-from-file", // Placeholder - actual key is in identities
		}
		// If recipient not provided, we can't derive it from the identity interface
		if recipient == "" {
			return fmt.Errorf("recipient must be provided when using existing key file")
		} else {
			keyPair.Recipient = recipient
		}
	} else if recipient != "" {
		// Only recipient provided (for shared stores)
		keyPair = &age.Key{
			Recipient: recipient,
		}
	} else {
		// Generate new key pair
		var err error
		keyPair, err = age.GenerateKeyPair()
		if err != nil {
			return fmt.Errorf("failed to generate age key: %w", err)
		}
		cmd.Printf("Generated new age key pair:\n")
		cmd.Printf("Recipient (public key): %s\n", keyPair.Recipient)
		cmd.Printf("Identity (private key): %s\n", keyPair.Identity)
		cmd.Printf("\nSave your private key securely!\n")
	}

	// Get store name
	storeName := cmd.Flag("store").Value.String()
	if storeName == "" {
		storeName = "personal"
	}

	// Create store directory
	storeDir := filepath.Join(configDir, "stores", storeName)
	if err := os.MkdirAll(storeDir, 0700); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	// Create .recipients file
	recipientsFile := filepath.Join(storeDir, ".recipients")
	recipientsData := fmt.Sprintf("# Age recipients for store '%s'\n%s\n", storeName, keyPair.Recipient)
	if err := os.WriteFile(recipientsFile, []byte(recipientsData), 0600); err != nil {
		return fmt.Errorf("failed to create recipients file: %w", err)
	}

	// Create or update config
	cfg := &config.Config{
		DefaultStore: storeName,
		Stores: map[string]config.StoreConfig{
			storeName: {
				Path:       storeDir,
				Recipients: []string{keyPair.Recipient},
			},
		},
		AgeKeyPath: filepath.Join(configDir, "age-key.txt"),
	}

	// Load existing config if it exists
	configFile := filepath.Join(configDir, "config.yaml")
	if data, err := os.ReadFile(configFile); err == nil {
		existingCfg := &config.Config{}
		if err := yaml.Unmarshal(data, existingCfg); err == nil {
			// Merge with existing config
			if existingCfg.Stores == nil {
				existingCfg.Stores = make(map[string]config.StoreConfig)
			}
			existingCfg.Stores[storeName] = cfg.Stores[storeName]
			if existingCfg.DefaultStore == "" {
				existingCfg.DefaultStore = storeName
			}
			cfg = existingCfg
		}
	}

	// Save config
	configData, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(configFile, configData, 0600); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Save private key if generated
	if keyPair.Identity != "" && ageKeyPath == "" {
		keyData := fmt.Sprintf("# Age private key for pf password manager\n# Public key: %s\n%s\n", 
			keyPair.Recipient, keyPair.Identity)
		if err := os.WriteFile(cfg.AgeKeyPath, []byte(keyData), 0600); err != nil {
			return fmt.Errorf("failed to save age key: %w", err)
		}
		cmd.Printf("\nPrivate key saved to: %s\n", cfg.AgeKeyPath)
	}

	cmd.Printf("\nPassword store '%s' initialized successfully!\n", storeName)
	cmd.Printf("Store location: %s\n", storeDir)
	cmd.Printf("Recipients file: %s\n", recipientsFile)

	return nil
}