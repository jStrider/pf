package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"pf/internal/age"
	"pf/internal/config"
)

// NewAgeCommand creates the age command
func NewAgeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "age",
		Short: "Manage age encryption keys",
		Long:  `Generate and manage age encryption keys`,
	}

	cmd.AddCommand(
		newAgeGenerateCommand(),
		newAgeExportCommand(),
		newAgeImportCommand(),
	)

	return cmd
}

func newAgeGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a new age key pair",
		RunE:  runAgeGenerate,
	}

	cmd.Flags().String("output", "", "Output file for private key")

	return cmd
}

func newAgeExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export public key (recipient)",
		RunE:  runAgeExport,
	}
}

func newAgeImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [file]",
		Short: "Import age private key",
		Args:  cobra.ExactArgs(1),
		RunE:  runAgeImport,
	}

	return cmd
}

func runAgeGenerate(cmd *cobra.Command, args []string) error {
	// Generate key pair
	keyPair, err := age.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Display key information
	cmd.Println("Generated age key pair:")
	cmd.Printf("Public key (recipient): %s\n", keyPair.Recipient)
	cmd.Printf("Private key (identity): %s\n", keyPair.Identity)

	// Save to file if requested
	output := cmd.Flag("output").Value.String()
	if output != "" {
		// Ensure directory exists
		dir := filepath.Dir(output)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Write key file
		keyData := fmt.Sprintf("# Age private key for pf password manager\n# Public key: %s\n%s\n",
			keyPair.Recipient, keyPair.Identity)
		if err := os.WriteFile(output, []byte(keyData), 0600); err != nil {
			return fmt.Errorf("failed to write key file: %w", err)
		}

		cmd.Printf("\nPrivate key saved to: %s\n", output)
		cmd.Println("Keep this file secure!")
	} else {
		cmd.Println("\nWARNING: Private key not saved to file.")
		cmd.Println("Save it securely or it will be lost!")
	}

	return nil
}

func runAgeExport(cmd *cobra.Command, args []string) error {
	// Load config to get key path
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load identity file
	identities, err := age.LoadIdentityFile(cfg.AgeKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load age key: %w", err)
	}

	if len(identities) == 0 {
		return fmt.Errorf("no identities found in key file")
	}

	// Export recipients
	cmd.Println("Age recipients (public keys):")
	cmd.Println("Note: Recipients cannot be derived from loaded identity files")
	cmd.Println("Please check your .recipients file or original key generation output")

	return nil
}

func runAgeImport(cmd *cobra.Command, args []string) error {
	keyFile := args[0]

	// Verify file exists
	if _, err := os.Stat(keyFile); err != nil {
		return fmt.Errorf("key file not found: %w", err)
	}

	// Load and validate key
	identities, err := age.LoadIdentityFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to load key file: %w", err)
	}

	if len(identities) == 0 {
		return fmt.Errorf("no valid identities found in file")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Copy key file to config directory
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("failed to read key file: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(cfg.AgeKeyPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write key file
	if err := os.WriteFile(cfg.AgeKeyPath, keyData, 0600); err != nil {
		return fmt.Errorf("failed to save key file: %w", err)
	}

	cmd.Printf("Age key imported successfully to: %s\n", cfg.AgeKeyPath)

	// Show imported recipients
	cmd.Println("\nAge key imported. Recipients cannot be shown from identity file.")
	cmd.Println("Save your recipient (public key) separately when generating keys.")

	return nil
}