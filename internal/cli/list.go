package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"pf/internal/config"
	"pf/internal/store"
)

// NewListCommand creates the list command
func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all password keys",
		Long:  `List all password keys in the store`,
		RunE:  runList,
		Aliases: []string{"ls"},
	}

	cmd.Flags().String("store", "", "Store name")
	cmd.Flags().Bool("tree", false, "Display as tree")

	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
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

	// Get all keys
	keys, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	if len(keys) == 0 {
		cmd.Printf("No passwords stored in '%s'\n", storeName)
		return nil
	}

	// Display as tree or list
	showTree, _ := cmd.Flags().GetBool("tree")
	if showTree {
		cmd.Printf("Password store '%s':\n", storeName)
		displayTree(cmd, keys)
	} else {
		cmd.Printf("Password store '%s':\n", storeName)
		for _, key := range keys {
			cmd.Printf("  %s\n", key)
		}
	}

	return nil
}

type treeNode struct {
	children map[string]*treeNode
	isLeaf   bool
}

func displayTree(cmd *cobra.Command, keys []string) {
	// Build tree
	root := &treeNode{children: make(map[string]*treeNode)}
	for _, key := range keys {
		parts := strings.Split(key, "/")
		current := root
		for _, part := range parts {
			if current.children[part] == nil {
				current.children[part] = &treeNode{children: make(map[string]*treeNode)}
			}
			current = current.children[part]
		}
		current.isLeaf = true
	}

	// Print tree
	printNode(cmd, root, "", "", true)
}

func printNode(cmd *cobra.Command, n *treeNode, prefix, name string, isLast bool) {
	if name != "" {
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		cmd.Printf("%s%s%s", prefix, connector, name)
		if n.isLeaf && len(n.children) == 0 {
			cmd.Printf("\n")
		} else {
			cmd.Printf("/\n")
		}
	}

	// Get sorted children
	childNames := make([]string, 0, len(n.children))
	for name := range n.children {
		childNames = append(childNames, name)
	}
	// Sort them
	for i := 0; i < len(childNames); i++ {
		for j := i + 1; j < len(childNames); j++ {
			if childNames[i] > childNames[j] {
				childNames[i], childNames[j] = childNames[j], childNames[i]
			}
		}
	}

	// Print children
	for i, childName := range childNames {
		isLastChild := i == len(childNames)-1
		childPrefix := prefix
		if name != "" {
			if isLast {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
		}
		printNode(cmd, n.children[childName], childPrefix, childName, isLastChild)
	}
}