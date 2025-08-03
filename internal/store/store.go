package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"filippo.io/age"

	pfage "pf/internal/age"
	"pf/internal/audit"
)

// Store represents a password store
type Store struct {
	path       string
	recipients []string
	identities []age.Identity
	auditor    *audit.Logger
}

// Entry represents a password entry with versioning
type Entry struct {
	Key      string    `yaml:"key"`
	Versions []Version `yaml:"versions"`
}

// Version represents a single version of a password
type Version struct {
	Version   int    `yaml:"version"`
	Password  string `yaml:"password"`
	Timestamp int64  `yaml:"timestamp"`
	Author    string `yaml:"author,omitempty"`
	Message   string `yaml:"message,omitempty"`
}

// New creates a new Store instance
func New(path, ageKeyPath string) (*Store, error) {
	// Load recipients from .recipients file
	recipientsFile := filepath.Join(path, ".recipients")
	recipients, err := loadRecipients(recipientsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load recipients: %w", err)
	}

	// Load identities
	var identities []age.Identity
	if ageKeyPath != "" {
		identities, err = pfage.LoadIdentityFile(ageKeyPath)
		if err != nil {
			// Non-fatal: might be a shared store
			identities = []age.Identity{}
		}
	}

	// Initialize auditor
	auditor := audit.New(filepath.Join(path, ".audit.log"))

	return &Store{
		path:       path,
		recipients: recipients,
		identities: identities,
		auditor:    auditor,
	}, nil
}

// Get retrieves a password by key
func (s *Store) Get(key string, version int) (string, error) {
	// Log audit event
	s.auditor.Log(audit.EventAccess, key, "")

	// Load entry
	entry, err := s.loadEntry(key)
	if err != nil {
		return "", err
	}

	// Get requested version
	if version <= 0 || version > len(entry.Versions) {
		// Get latest version
		version = len(entry.Versions)
	}

	v := entry.Versions[version-1]

	// Decrypt password
	password, err := pfage.Decrypt(v.Password, s.identities)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt password: %w", err)
	}

	return password, nil
}

// Put stores a new password or updates an existing one
func (s *Store) Put(key, password, message string) error {
	// Log audit event
	s.auditor.Log(audit.EventModify, key, message)

	// Encrypt password
	encrypted, err := pfage.Encrypt(password, s.recipients)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Load or create entry
	entry, err := s.loadEntry(key)
	if err != nil {
		// Create new entry
		entry = &Entry{
			Key:      key,
			Versions: []Version{},
		}
	}

	// Add new version
	newVersion := Version{
		Version:   len(entry.Versions) + 1,
		Password:  encrypted,
		Timestamp: time.Now().Unix(),
		Author:    os.Getenv("USER"),
		Message:   message,
	}
	entry.Versions = append(entry.Versions, newVersion)

	// Save entry
	return s.saveEntry(entry)
}

// Delete removes a password entry
func (s *Store) Delete(key string) error {
	// Log audit event
	s.auditor.Log(audit.EventDelete, key, "")

	// Remove entry file
	entryPath := s.getEntryPath(key)
	if err := os.Remove(entryPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("password '%s' not found", key)
		}
		return fmt.Errorf("failed to delete password: %w", err)
	}

	return nil
}

// GetHistory retrieves the version history of a password
func (s *Store) GetHistory(key string, limit int) ([]Version, error) {
	// Load entry
	entry, err := s.loadEntry(key)
	if err != nil {
		return nil, err
	}

	// Get versions in reverse order (newest first)
	versions := make([]Version, len(entry.Versions))
	copy(versions, entry.Versions)
	
	// Reverse the slice
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	// Apply limit
	if limit > 0 && limit < len(versions) {
		versions = versions[:limit]
	}

	// Clear passwords from history
	for i := range versions {
		versions[i].Password = ""
	}

	return versions, nil
}

// List returns all password keys in the store
func (s *Store) List() ([]string, error) {
	var keys []string
	
	// Walk through the directory tree
	err := filepath.Walk(s.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Only process .yaml files
		if strings.HasSuffix(info.Name(), ".yaml") {
			// Get relative path from store root
			relPath, err := filepath.Rel(s.path, path)
			if err != nil {
				return err
			}
			
			// Remove .yaml extension to get the key
			key := strings.TrimSuffix(relPath, ".yaml")
			
			// Skip hidden files (starting with .)
			if !strings.HasPrefix(filepath.Base(key), ".") && !strings.Contains(key, "/.") {
				// Convert file path separators to forward slashes for consistency
				key = filepath.ToSlash(key)
				keys = append(keys, key)
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to walk store directory: %w", err)
	}

	sort.Strings(keys)
	return keys, nil
}

// Helper functions

func (s *Store) getEntryPath(key string) string {
	// Support hierarchical structure
	// Convert the key to a file path, ensuring the .yaml extension
	// For example: "github.com/user" becomes "github.com/user.yaml"
	return filepath.Join(s.path, key+".yaml")
}

func (s *Store) loadEntry(key string) (*Entry, error) {
	entryPath := s.getEntryPath(key)
	
	data, err := os.ReadFile(entryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("password '%s' not found", key)
		}
		return nil, fmt.Errorf("failed to read entry: %w", err)
	}

	var entry Entry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse entry: %w", err)
	}

	return &entry, nil
}

func (s *Store) saveEntry(entry *Entry) error {
	entryPath := s.getEntryPath(entry.Key)
	
	// Create parent directories if they don't exist
	parentDir := filepath.Dir(entryPath)
	if err := os.MkdirAll(parentDir, 0700); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}
	
	data, err := yaml.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	if err := os.WriteFile(entryPath, data, 0600); err != nil {
		return fmt.Errorf("failed to save entry: %w", err)
	}

	return nil
}

func loadRecipients(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var recipients []string
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Validate recipient
		if _, err := pfage.ParseRecipient(line); err == nil {
			recipients = append(recipients, line)
		}
	}

	return recipients, nil
}