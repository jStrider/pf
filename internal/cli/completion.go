package cli

import (
	"sort"
	"strings"
	
	"github.com/spf13/cobra"
	
	"pf/internal/config"
	"pf/internal/store"
)

// passwordKeyCompletion provides completion for password keys
func passwordKeyCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Don't provide completion if we already have an argument
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get store name from flag or use default
	storeName, _ := cmd.Flags().GetString("store")
	if storeName == "" {
		storeName = cfg.DefaultStore
	}

	// Get store config
	storeConfig, ok := cfg.Stores[storeName]
	if !ok {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Initialize store
	s, err := store.New(storeConfig.Path, cfg.AgeKeyPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get all keys
	keys, err := s.List()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Build a tree structure for better path completion
	type node struct {
		children map[string]*node
		isLeaf   bool
		fullPath string
	}
	
	root := &node{children: make(map[string]*node)}
	
	// Build tree from keys
	for _, key := range keys {
		parts := strings.Split(key, "/")
		current := root
		for i, part := range parts {
			if current.children[part] == nil {
				current.children[part] = &node{
					children: make(map[string]*node),
					fullPath: strings.Join(parts[:i+1], "/"),
				}
			}
			current = current.children[part]
		}
		current.isLeaf = true
		current.fullPath = key
	}
	
	// Generate completions based on current input
	var completions []string
	
	if toComplete == "" {
		// Show top-level entries and directories
		for name, child := range root.children {
			if child.isLeaf && len(child.children) == 0 {
				completions = append(completions, name)
			} else {
				// For directories, add with slash to show it's a directory
				completions = append(completions, name+"/")
			}
		}
	} else if strings.HasSuffix(toComplete, "/") {
		// User typed a complete directory path with trailing slash
		// Show contents of that directory
		parts := strings.Split(strings.TrimSuffix(toComplete, "/"), "/")
		current := root
		
		// Navigate to the directory
		for _, part := range parts {
			if next, ok := current.children[part]; ok {
				current = next
			} else {
				// Path doesn't exist
				return completions, cobra.ShellCompDirectiveNoFileComp
			}
		}
		
		// Show children of current directory
		prefix := toComplete
		for name, child := range current.children {
			if child.isLeaf && len(child.children) == 0 {
				completions = append(completions, prefix+name)
			} else {
				// Add slash for directories
				completions = append(completions, prefix+name+"/")
			}
		}
	} else {
		// Partial path without trailing slash
		parts := strings.Split(toComplete, "/")
		current := root
		
		// Navigate to parent of last component
		for i := 0; i < len(parts)-1; i++ {
			if next, ok := current.children[parts[i]]; ok {
				current = next
			} else {
				// Path doesn't exist
				return completions, cobra.ShellCompDirectiveNoFileComp
			}
		}
		
		// Find matches for the last component
		lastPart := parts[len(parts)-1]
		prefix := ""
		if len(parts) > 1 {
			prefix = strings.Join(parts[:len(parts)-1], "/") + "/"
		}
		
		for name, child := range current.children {
			if strings.HasPrefix(name, lastPart) {
				if child.isLeaf && len(child.children) == 0 {
					completions = append(completions, prefix+name)
				} else {
					// For directories, add with slash
					completions = append(completions, prefix+name+"/")
				}
			}
		}
	}
	
	sort.Strings(completions)
	
	// Always use NoSpace directive to allow continuous path completion
	// This prevents adding a space after selecting a directory
	return completions, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}

// storeNameCompletion provides completion for store names
func storeNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Get all store names
	var stores []string
	for name := range cfg.Stores {
		stores = append(stores, name)
	}

	return stores, cobra.ShellCompDirectiveNoFileComp
}