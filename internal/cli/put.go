package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"pf/internal/config"
	"pf/internal/store"
)

// NewPutCommand creates the put command
func NewPutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put [key]",
		Short: "Store a password",
		Long:  `Store a new password or update an existing one`,
		Args:  cobra.ExactArgs(1),
		RunE:  runPut,
		ValidArgsFunction: passwordKeyCompletion,
	}

	cmd.Flags().String("store", "", "Store name")
	cmd.Flags().Bool("multiline", false, "Enable multiline input")
	cmd.Flags().String("message", "", "Version message")

	return cmd
}

func runPut(cmd *cobra.Command, args []string) error {
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

	// Get password input
	var password string
	multiline, _ := cmd.Flags().GetBool("multiline")

	if multiline {
		cmd.Println("Enter password (press Ctrl+D when done):")
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		password = strings.Join(lines, "\n")
	} else {
		// Check if input is from pipe
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Input from pipe
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				password = scanner.Text()
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
		} else {
			// Interactive input
			fmt.Print("Enter password: ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
			fmt.Println()
			
			fmt.Print("Confirm password: ")
			byteConfirm, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("failed to read password confirmation: %w", err)
			}
			fmt.Println()
			
			if string(bytePassword) != string(byteConfirm) {
				return fmt.Errorf("passwords do not match")
			}
			
			password = string(bytePassword)
		}
	}

	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Get version message
	message := cmd.Flag("message").Value.String()

	// Store password
	if err := s.Put(key, password, message); err != nil {
		return fmt.Errorf("failed to store password: %w", err)
	}

	cmd.Printf("Password stored for '%s' in store '%s'\n", key, storeName)
	return nil
}