package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"pf/internal/config"
	"pf/internal/store"
)

// NewHistoryCommand creates the history command
func NewHistoryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history [key]",
		Short: "Show password history",
		Long:  `Display the version history of a password`,
		Args:  cobra.ExactArgs(1),
		RunE:  runHistory,
		ValidArgsFunction: passwordKeyCompletion,
	}

	cmd.Flags().String("store", "", "Store name")
	cmd.Flags().Int("limit", 10, "Maximum number of versions to show")

	return cmd
}

func runHistory(cmd *cobra.Command, args []string) error {
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

	// Get history
	limit, _ := cmd.Flags().GetInt("limit")
	history, err := s.GetHistory(key, limit)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	if len(history) == 0 {
		cmd.Printf("No history found for '%s'\n", key)
		return nil
	}

	// Display history
	cmd.Printf("History for '%s':\n\n", key)
	for i, version := range history {
		timestamp := time.Unix(version.Timestamp, 0).Format("2006-01-02 15:04:05")
		current := ""
		if i == 0 {
			current = " (current)"
		}
		
		cmd.Printf("Version %d%s - %s\n", version.Version, current, timestamp)
		if version.Message != "" {
			cmd.Printf("  Message: %s\n", version.Message)
		}
		if version.Author != "" {
			cmd.Printf("  Author: %s\n", version.Author)
		}
		cmd.Println()
	}

	return nil
}