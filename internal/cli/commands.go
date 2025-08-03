package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pf",
		Short: "PF - Password manager with age encryption",
		Long: `PF is a secure password manager that uses age encryption.

It provides a simple interface for storing and retrieving passwords,
with features like versioning, multiple stores, and audit logging.`,
	}

	cmd.AddCommand(
		NewInitCommand(),
		NewGetCommand(),
		NewPutCommand(),
		NewDeleteCommand(),
		NewListCommand(),
		NewHistoryCommand(),
		NewRollbackCommand(),
		NewStoreCommand(),
		NewConfigCommand(),
		NewAgeCommand(),
	)

	return cmd
}