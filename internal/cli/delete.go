package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"pf/internal/config"
	"pf/internal/store"
)

// NewDeleteCommand creates the delete command
func NewDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [key]",
		Short: "Delete a password",
		Long:  `Delete a password from the store`,
		Args:  cobra.ExactArgs(1),
		RunE:  runDelete,
		ValidArgsFunction: passwordKeyCompletion,
	}

	cmd.Flags().String("store", "", "Store name")
	cmd.Flags().Bool("force", false, "Skip confirmation")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	key := args[0]

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

	// Confirm deletion
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		fmt.Printf("Are you sure you want to delete '%s'? [y/N] ", key)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			cmd.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete password
	if err := s.Delete(key); err != nil {
		return fmt.Errorf("failed to delete password: %w", err)
	}

	cmd.Printf("Password '%s' deleted from store '%s'\n", key, storeName)
	return nil
}