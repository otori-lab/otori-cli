# Otori CLI

Outil CLI pour déployer des honeypots Cowrie locaux. Projet de Fin d'Études (PFE) - ECE Paris.

## Installation

```bash
git clone https://github.com/otori-lab/otori-cli.git
cd otori-cli
make build
```

Le binaire est généré dans `bin/otori`.

## Utilisation rapide

```bash
# Créer un profil
./bin/otori init -t classic -p mon-honeypot -s srv-prod-01 -u root,admin

# Déployer le honeypot
./bin/otori deploy -p mon-honeypot

# Vérifier le statut
./bin/otori status

# Arrêter
./bin/otori stop -p mon-honeypot
```

## Commandes

| Commande | Description |
|----------|-------------|
| `init` | Crée un profil de honeypot |
| `deploy` | Déploie le honeypot via Docker |
| `status` | Affiche l'état des honeypots |
| `stop` | Arrête un honeypot |
| `profiles list` | Liste les profils |
| `profiles show` | Affiche les détails d'un profil |
| `profiles delete` | Supprime un profil |

Voir [internal/commands/README.md](internal/commands/README.md) pour la documentation détaillée.

## Architecture

```
~/.otori/profiles/{profile}/
├── {profile}.json      # Configuration du profil
├── cowrie.cfg          # Config Cowrie
├── userdb.txt          # Utilisateurs autorisés
├── docker-compose.yml  # Compose pour déploiement
└── honeyfs/            # Filesystem simulé
    ├── etc/
    ├── proc/
    └── ...
```

## Personnalisation du honeyfs

Après `init`, vous pouvez ajouter des fichiers dans le dossier `honeyfs/` du profil. Ils seront automatiquement ajoutés au filesystem du honeypot lors du `deploy`.

```bash
# Exemple : ajouter un fichier bait
echo "DB_PASS=secret123" > ~/.otori/profiles/mon-honeypot/honeyfs/var/www/.env

# Le fichier sera visible dans le honeypot après deploy
./bin/otori deploy -p mon-honeypot
```

## Prérequis

- Go 1.21+
- Docker & Docker Compose

## Licence

MIT
