package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"pf/internal/config"
	"pf/internal/store"
)

// NewRollbackCommand creates the rollback command
func NewRollbackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback [key] [version]",
		Short: "Rollback to a previous password version",
		Long:  `Restore a password to a previous version`,
		Args:  cobra.ExactArgs(2),
		RunE:  runRollback,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// First argument is the key
			if len(args) == 0 {
				return passwordKeyCompletion(cmd, args, toComplete)
			}
			// Second argument is the version number - no completion
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().String("store", "", "Store name")
	cmd.Flags().String("message", "", "Rollback message")

	return cmd
}

func runRollback(cmd *cobra.Command, args []string) error {
	key := args[0]
	version := args[1]

	// Parse version number
	var versionNum int
	if _, err := fmt.Sscanf(version, "%d", &versionNum); err != nil {
		return fmt.Errorf("invalid version number: %s", version)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get store
	storeName := cmd.Flag("store").Value.String()
	if storeName == "" {
		storeName = cfg.DefaultStore
	}

	storeConfig, ok := cfg.Stores[storeName]
	if !ok {
		return fmt.Errorf("store '%s' not found", storeName)
	}

	// Initialize store
	s, err := store.New(storeConfig.Path, cfg.AgeKeyPath)
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	// Get the old version
	oldPassword, err := s.Get(key, versionNum)
	if err != nil {
		return fmt.Errorf("failed to get version %d: %w", versionNum, err)
	}

	// Create rollback message
	message := cmd.Flag("message").Value.String()
	if message == "" {
		message = fmt.Sprintf("Rollback to version %d", versionNum)
	}

	// Store as new version
	if err := s.Put(key, oldPassword, message); err != nil {
		return fmt.Errorf("failed to rollback: %w", err)
	}

	cmd.Printf("Successfully rolled back '%s' to version %d\n", key, versionNum)
	return nil
}