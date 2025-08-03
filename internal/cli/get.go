package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"pf/internal/config"
	"pf/internal/store"
)

// NewGetCommand creates the get command
func NewGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a password",
		Long:  `Retrieve a password from the store`,
		Args:  cobra.ExactArgs(1),
		RunE:  runGet,
		ValidArgsFunction: passwordKeyCompletion,
	}

	cmd.Flags().String("store", "", "Store name")
	cmd.Flags().Bool("clip", false, "Copy password to clipboard")
	cmd.Flags().Duration("clip-time", 45*time.Second, "Clipboard clearing time")
	cmd.Flags().Int("version", 0, "Get specific version (0 = latest)")

	return cmd
}

func runGet(cmd *cobra.Command, args []string) error {
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

	// Get password
	version, _ := cmd.Flags().GetInt("version")
	password, err := s.Get(key, version)
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	// Output or copy to clipboard
	if clip, _ := cmd.Flags().GetBool("clip"); clip {
		if err := clipboard.WriteAll(password); err != nil {
			return fmt.Errorf("failed to copy to clipboard: %w", err)
		}
		
		clipTime, _ := cmd.Flags().GetDuration("clip-time")
		cmd.Printf("Password for '%s' copied to clipboard. Will clear in %s.\n", key, clipTime)
		
		// Clear clipboard after timeout
		go func() {
			time.Sleep(clipTime)
			clipboard.WriteAll("")
		}()
	} else {
		fmt.Fprint(os.Stdout, password)
		if password[len(password)-1] != '\n' {
			fmt.Fprintln(os.Stdout)
		}
	}

	return nil
}