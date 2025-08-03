# pf - CLI Password Manager

> 🔐 Un gestionnaire de mots de passe en ligne de commande sécurisé utilisant age

## 🚀 Installation rapide

```bash
# Prérequis : Go 1.21+
git clone https://github.com/julienrenaudperso/pf
cd pf
make build
make install  # Installation globale (nécessite sudo)
```

## 📋 Commandes principales

### Gestion des mots de passe

| Commande | Description | Exemple |
|----------|-------------|---------|
| `init [path]` | Initialise un nouveau store | `pf init ~/.passwords` |
| `get <path>` | Récupère un mot de passe | `pf get email/gmail --clip` |
| `put <path>` | Stocke/met à jour un mot de passe | `pf put email/gmail --editor` |
| `delete <path>` | Supprime un mot de passe | `pf delete email/gmail` |
| `history <path>` | Affiche l'historique des versions | `pf history email/gmail` |
| `rollback <path> <n>` | Restaure la version n | `pf rollback email/gmail 2` |

### Gestion des stores

| Commande | Description |
|----------|-------------|
| `store show` | Affiche le store par défaut |
| `store list` | Liste tous les stores configurés |
| `store setup <name> <path>` | Crée un nouveau store |
| `store ls [store]` | Liste les mots de passe d'un store |

### Configuration et age

| Commande | Description |
|----------|-------------|
| `config show` | Affiche la configuration |
| `config edit` | Édite la configuration |
| `config verify` | Vérifie la configuration |
| `age generate` | Génère une nouvelle paire de clés age |
| `age list-keys` | Liste les destinataires age |
| `age add-key <recipient>` | Ajoute un destinataire age |
| `age import <file>` | Importe une identité age |

## 🔧 Configuration

**Fichier**: `~/.config/pf/config.yaml`

```yaml
global:
    default_store: mystore_1         # Store par défaut
    editor: nano                     # Éditeur de texte
    clipboard_timeout: 30s           # Durée avant effacement du presse-papier
    default_get_mode: clip           # Mode par défaut: "clip" ou "show"

stores:
    mystore_1:
        path: /home/user/.pf-store   # Chemin du store
        editor: nano                 # Éditeur spécifique (optionnel)
        audit: /home/user/.pf-store/logs

audit:
    enabled: true                    # Active l'audit
    log_failed_attempts: true        # Journalise les échecs
    alert_on_suspicious_activity: true
    retention_days: 365              # Durée de conservation des logs
```

## 📁 Structure d'un store

```
~/.pf-store/
├── .recipients         # Clés publiques age (1 par ligne)
├── data/              # Mots de passe chiffrés
│   └── email/
│       └── gmail.yaml
├── metadata/          # Métadonnées des versions
│   └── email/
│       └── gmail.yaml
└── logs/              # Journaux d'audit
    └── 2025-08-03.log
```

### Format des fichiers

**data/email/gmail.yaml** (chiffré age)
```yaml
name: gmail
versions:
  - version: 1
    encrypted_value: |
      -----BEGIN AGE ENCRYPTED FILE-----
      [contenu chiffré]
      -----END AGE ENCRYPTED FILE-----
```

**metadata/email/gmail.yaml**
```yaml
name: gmail
versions:
  - version: 1
    created_at: 2025-08-03T12:00:00Z
    author: julien
```

**Format des logs**
```
<timestamp>  <action>  <actor>  <path>  <version>
2025-08-03T12:00:00Z  put  julien  email/gmail  1
```

## 🔒 Sécurité

- **Chiffrement**: age (moderne et simple) avec clés publiques dans `.recipients`
- **Clé privée**: Stockée dans `~/.config/age/key.txt` (protégée 0600)
- **Permissions**: 
  - Store: `0700` (lecture/écriture/exécution propriétaire)
  - Fichiers: `0600` (lecture/écriture propriétaire)
- **Presse-papier**: Effacement automatique après timeout
- **Audit**: Journalisation complète des actions

## 💡 Exemples d'utilisation

### Démarrage rapide
```bash
# Générer une clé age (si vous n'en avez pas)
pf age generate

# Initialiser un store
pf init
# Le système propose de générer une clé si aucune n'est fournie

# Ajouter un mot de passe
pf put email/gmail
# Saisir le mot de passe (masqué)

# Récupérer dans le presse-papier
pf get email/gmail

# Voir l'historique
pf history email/gmail
```

### Organisation hiérarchique
```bash
pf put personal/email/gmail
pf put personal/social/twitter
pf put work/vpn
pf put work/gitlab
pf put banking/mybank
```

### Multi-stores
```bash
# Créer un store "work"
pf store setup work ~/work-passwords

# Utiliser un store spécifique
pf get gitlab --store work
```

### Options avancées
```bash
# Mode éditeur pour mots de passe complexes
pf put ssh/server --editor

# Afficher sur stdout au lieu du presse-papier
pf get ssh/server --show

# Récupérer une version spécifique
pf get email/gmail --version 2

# Supprimer sans confirmation
pf delete old/account --force
```

## 🛠️ Développement

```bash
make build         # Compiler
make test          # Tests unitaires
make fmt           # Formater le code
make lint          # Vérifier le code
make clean         # Nettoyer
```


## 📄 Licence

MIT License - Voir le fichier [LICENSE](LICENSE) pour plus de détails.