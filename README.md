# pf - CLI Password Manager

> 🔐 A secure command-line password manager using age encryption

## 🚀 Quick Installation

```bash
# Prerequisites: Go 1.21+
git clone https://github.com/julienrenaudperso/pf
cd pf
make build
make install  # Global installation
```

## 📋 Main Commands

### Password Management

| Command | Description | Example |
|---------|-------------|---------|
| `init` | Initialize a new store | `pf init` |
| `get <key>` | Retrieve a password | `pf get email/gmail --clip` |
| `put <key>` | Store/update a password | `pf put email/gmail` |
| `delete <key>` | Delete a password | `pf delete email/gmail` |
| `list` | List all passwords | `pf list --tree` |
| `history <key>` | Show version history | `pf history email/gmail` |
| `rollback <key> <n>` | Restore version n | `pf rollback email/gmail 2` |

### Store Management

| Command | Description |
|---------|-------------|
| `store list` | List all configured stores |
| `store add <name>` | Create a new store |
| `store remove <name>` | Remove a store |
| `store set-default <name>` | Set default store |

### Configuration and age

| Command | Description |
|---------|-------------|
| `config show` | Display configuration |
| `config get <key>` | Get a configuration value |
| `config set <key> <value>` | Set a configuration value |
| `age generate` | Generate a new age key pair |
| `age export` | Export public key (recipient) |
| `age import <file>` | Import age private key |

## 🔧 Configuration

**File**: `~/.pf/config.yaml`

```yaml
default_store: personal              # Default store
stores:
  personal:
    path: ~/.pf/stores/personal     # Store path
    recipients:                     # age recipients
      - age1abc...xyz
age_key_path: ~/.pf/age-key.txt    # Private key path
clipboard_timeout: 45s              # Clipboard clearing timeout
audit_log: true                     # Enable audit logging
```

## 📁 Store Structure

```
~/.pf/stores/personal/
├── .recipients         # age public keys (one per line)
├── .audit.log         # Audit log (if enabled)
├── email/
│   └── gmail.yaml     # Encrypted password file
├── banking/
│   └── mybank.yaml
└── work/
    ├── gitlab.yaml
    └── vpn.yaml
```

### File Format

**Password file (e.g., email/gmail.yaml)**
```yaml
key: email/gmail
versions:
  - version: 1
    password: |
      -----BEGIN AGE ENCRYPTED FILE-----
      [encrypted content]
      -----END AGE ENCRYPTED FILE-----
    timestamp: 1722686400
    author: julien
    message: "Initial password"
```

**Audit log format**
```
2025-08-03T12:00:00Z | julien | ACCESS | email/gmail | 
2025-08-03T12:01:00Z | julien | MODIFY | email/gmail | Updated password
```

## 🔒 Security

- **Encryption**: age (modern and simple) with public keys in `.recipients`
- **Private key**: Stored securely with 0600 permissions
- **Permissions**: 
  - Stores: `0700` (owner read/write/execute)
  - Files: `0600` (owner read/write)
- **Clipboard**: Automatic clearing after timeout
- **Audit**: Complete action logging

## 💡 Usage Examples

### Quick Start
```bash
# Initialize a store (generates age key if needed)
pf init

# Add a password
pf put email/gmail
# Enter password (hidden input)

# Retrieve to clipboard
pf get email/gmail

# View history
pf history email/gmail
```

### Hierarchical Organization
```bash
pf put personal/email/gmail
pf put personal/social/twitter
pf put work/vpn
pf put work/gitlab
pf put banking/mybank

# List with tree view
pf list --tree
```

### Multi-store Usage
```bash
# Create a "work" store
pf store add work

# Use a specific store
pf get gitlab --store work

# Set default store
pf store set-default work
```

### Advanced Options
```bash
# Multiline passwords
pf put ssh/server --multiline

# Display on stdout instead of clipboard
pf get ssh/server

# Retrieve with clipboard timeout
pf get email/gmail --clip-time 60s

# Get specific version
pf get email/gmail --version 2

# Delete without confirmation
pf delete old/account --force
```

### Shell Completion

The password manager supports intelligent shell completion:

```bash
# Install completions (done by make install)
pf completion zsh > ~/.zsh/completions/_pf

# Tab completion examples
pf get <TAB>              # Shows all keys
pf get aws/<TAB>          # Shows: dev/ prod/
pf get aws/prod/<TAB>     # Shows passwords in aws/prod/
```

## 🛠️ Development

```bash
make build         # Build the binary
make test          # Run unit tests
make lint          # Lint the code
make install       # Install with completions
make clean         # Clean build artifacts
```

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details.