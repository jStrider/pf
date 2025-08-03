package age

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"filippo.io/age"
	"filippo.io/age/armor"
)

// Encrypt encrypts data for the given age recipients
func Encrypt(data string, recipients []string) (string, error) {
	if len(recipients) == 0 {
		return "", fmt.Errorf("no recipients specified")
	}

	// Parse recipients
	var ageRecipients []age.Recipient
	for _, recipient := range recipients {
		r, err := age.ParseX25519Recipient(recipient)
		if err != nil {
			return "", fmt.Errorf("invalid recipient %s: %w", recipient, err)
		}
		ageRecipients = append(ageRecipients, r)
	}

	// Encrypt data
	var encrypted bytes.Buffer
	armorWriter := armor.NewWriter(&encrypted)
	
	w, err := age.Encrypt(armorWriter, ageRecipients...)
	if err != nil {
		return "", fmt.Errorf("failed to create encryptor: %w", err)
	}

	if _, err := io.WriteString(w, data); err != nil {
		return "", fmt.Errorf("failed to write data: %w", err)
	}

	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close encryptor: %w", err)
	}
	
	if err := armorWriter.Close(); err != nil {
		return "", fmt.Errorf("failed to close armor writer: %w", err)
	}

	return encrypted.String(), nil
}

// Decrypt decrypts age encrypted data
func Decrypt(encrypted string, identities []age.Identity) (string, error) {
	if len(identities) == 0 {
		return "", fmt.Errorf("no identities provided")
	}

	// Create armor reader
	armorReader := armor.NewReader(strings.NewReader(encrypted))
	
	r, err := age.Decrypt(armorReader, identities...)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	var decrypted bytes.Buffer
	if _, err := io.Copy(&decrypted, r); err != nil {
		return "", fmt.Errorf("failed to read decrypted data: %w", err)
	}

	return decrypted.String(), nil
}

// GenerateKey generates a new age X25519 key pair
func GenerateKey() (*age.X25519Identity, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("failed to generate identity: %w", err)
	}
	return identity, nil
}

// ParseIdentity parses an age identity from string
func ParseIdentity(s string) (age.Identity, error) {
	identity, err := age.ParseX25519Identity(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
	}
	return identity, nil
}

// ParseRecipient parses an age recipient from string
func ParseRecipient(s string) (age.Recipient, error) {
	recipient, err := age.ParseX25519Recipient(s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse recipient: %w", err)
	}
	return recipient, nil
}

// Key represents an age key pair
type Key struct {
	Identity  string // Private key
	Recipient string // Public key
}

// GenerateKeyPair generates a new age key pair
func GenerateKeyPair() (*Key, error) {
	identity, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	return &Key{
		Identity:  identity.String(),
		Recipient: identity.Recipient().String(),
	}, nil
}

// LoadIdentityFile loads age identities from a file
func LoadIdentityFile(path string) ([]age.Identity, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read identity file: %w", err)
	}

	var identities []age.Identity
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		identity, err := ParseIdentity(line)
		if err != nil {
			// Skip invalid lines
			continue
		}
		identities = append(identities, identity)
	}

	if len(identities) == 0 {
		return nil, fmt.Errorf("no valid identities found in file")
	}

	return identities, nil
}