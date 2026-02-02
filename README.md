# Otori CLI

Outil CLI pour deployer et orchestrer des honeypots. Projet de Fin d'Etudes (PFE) - ECE Paris 2025.

## Installation

```bash
git clone https://github.com/otori-lab/otori-cli.git
cd otori-cli
make install
```

Le binaire est installe dans `~/.local/bin/otori`.

## Quick Start

```bash
# 1. Demarrer le monitoring (optionnel mais recommande)
otori monitoring start

# 2. Creer un profil de honeypot avec connexion au monitoring
otori init -t classic -p mon-honeypot -s srv-prod-01 -u root,admin \
  --monitoring-url "http://otori-monitoring:8000"

# 3. Deployer le honeypot
otori deploy -p mon-honeypot

# 4. Tester la connexion
ssh -p <port> root@localhost

# 5. Voir le dashboard
# Ouvrir http://localhost:8000
```

## Commandes

| Commande | Description |
|----------|-------------|
| `init` | Cree un profil de honeypot (classic ou ia) |
| `deploy` | Deploie le honeypot via Docker |
| `status` | Affiche l'etat des honeypots en cours |
| `stop` | Arrete un honeypot |
| `profiles list` | Liste les profils disponibles |
| `profiles show` | Affiche les details d'un profil |
| `profiles delete` | Supprime un profil |
| `monitoring start` | Demarre le serveur de monitoring |
| `monitoring stop` | Arrete le serveur de monitoring |
| `monitoring logs` | Affiche les logs du monitoring |

## Types de honeypots

### Classic (Cowrie)

Honeypot SSH/Telnet base sur Cowrie avec filesystem simule.

```bash
otori init -t classic -s prod-server -p my-honeypot
```

### IA (LLM)

Honeypot SSH avec reponses generees par IA (Ollama/Mistral).

```bash
otori init -t ia -s ai-server -p my-ai-honeypot
```

## Integration Monitoring

Le CLI integre automatiquement la connexion au monitoring :

```bash
# Deploiement local (meme machine)
otori init -t classic -s server01 --monitoring-url "http://otori-monitoring:8000"

# Deploiement distant (monitoring sur une autre machine)
otori init -t classic -s server01 --monitoring-url "http://192.168.1.100:8000"
```

Le reseau Docker `otori-network` est cree automatiquement pour permettre la communication entre les honeypots et le monitoring.

## Architecture

```
~/.otori/profiles/{profile}/
├── {profile}.json      # Configuration du profil
├── cowrie.cfg          # Config Cowrie (classic)
├── userdb.txt          # Utilisateurs autorises
├── docker-compose.yml  # Compose pour deploiement
├── shipper/            # Log shipper vers monitoring
└── honeyfs/            # Filesystem simule
    ├── etc/
    ├── home/
    ├── var/
    └── ...
```

## Personnalisation du honeyfs

Apres `init`, vous pouvez ajouter des fichiers "bait" dans le dossier `honeyfs/` :

```bash
# Ajouter un faux fichier de credentials
echo "DB_PASS=secret123" > ~/.otori/profiles/mon-honeypot/honeyfs/var/www/.env

# Deployer pour appliquer les changements
otori deploy -p mon-honeypot
```

## Flux de donnees

```
Attaquant SSH
     │
     ▼
┌─────────────┐
│  Honeypot   │ (Cowrie ou LLM)
│  Port 2222  │
└──────┬──────┘
       │ Logs JSON
       ▼
┌─────────────┐
│   Shipper   │ (Sidecar container)
└──────┬──────┘
       │ POST /ingest
       ▼
┌─────────────┐
│ Monitoring  │ (FastAPI + PostgreSQL)
│  Port 8000  │
└─────────────┘
       │
       ▼
   Dashboard (GeoIP, MITRE ATT&CK, Analytics)
```

## Prerequis

- Go 1.21+
- Docker & Docker Compose

## Licence

MIT
