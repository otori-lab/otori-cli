# Commands

Documentation détaillée des commandes Otori CLI.

## init

Crée un profil de honeypot.

```bash
# Mode interactif
otori init

# Mode non-interactif
otori init -t classic -p mon-profil -s srv-prod -u root,admin -c "Ma Company"
```

**Flags :**

| Flag | Court | Description |
|------|-------|-------------|
| `--type` | `-t` | Type de honeypot : `classic` ou `ia` |
| `--profile-name` | `-p` | Nom du profil (défaut: `default`) |
| `--server-name` | `-s` | Hostname du serveur simulé |
| `--company` | `-c` | Nom de l'organisation simulée |
| `--users` | `-u` | Liste d'utilisateurs séparés par virgule |

**Fichiers générés (type classic) :**
- `{profile}.json` - Configuration
- `cowrie.cfg` - Config Cowrie
- `userdb.txt` - Utilisateurs SSH autorisés
- `docker-compose.yml` - Déploiement Docker
- `honeyfs/` - Filesystem simulé

---

## deploy

Déploie un honeypot à partir d'un profil.

```bash
otori deploy -p mon-profil
otori deploy -p mon-profil -f  # Force recreate
```

**Flags :**

| Flag | Court | Description |
|------|-------|-------------|
| `--profile` | `-p` | Profil à déployer (défaut: `default`) |
| `--force` | `-f` | Force la recréation du container |

**Actions :**
1. Lance `docker compose up -d`
2. Scanne le `honeyfs/` pour détecter les fichiers customs
3. Met à jour le `fs.pickle` via fsctl (pour que `ls` voie les fichiers)
4. Restart le container

**Ports exposés :**
- `2222` - SSH
- `2223` - Telnet

---

## status

Affiche l'état des honeypots.

```bash
otori status              # Honeypots actifs
otori status -a           # Tous (y compris stoppés)
otori status -p mon-profil
otori status -j           # Sortie JSON
```

**Flags :**

| Flag | Court | Description |
|------|-------|-------------|
| `--profile` | `-p` | Filtrer par profil |
| `--all` | `-a` | Afficher tous les profils |
| `--json` | `-j` | Sortie JSON |

---

## stop

Arrête un honeypot.

```bash
otori stop -p mon-profil
otori stop -p mon-profil -f  # Force (arrêt immédiat)
```

**Flags :**

| Flag | Court | Description |
|------|-------|-------------|
| `--profile` | `-p` | Profil à arrêter (défaut: `default`) |
| `--force` | `-f` | Arrêt immédiat (timeout 0) |

---

## profiles

Gestion des profils.

```bash
otori profiles list              # Liste tous les profils
otori profiles show mon-profil   # Détails d'un profil
otori profiles delete mon-profil # Supprime un profil
```

---

## Fonctionnement du honeyfs

Le honeypot Cowrie utilise deux systèmes :

1. **fs.pickle** - Structure du filesystem (ce que `ls` affiche)
2. **honeyfs/** - Contenu des fichiers (ce que `cat` retourne)

Lors du `deploy`, Otori :
- Monte le dossier `honeyfs/` dans le container
- Détecte les fichiers qui n'existent pas dans le fs.pickle de base
- Les ajoute automatiquement via `fsctl`

Cela permet d'ajouter des fichiers "bait" personnalisés sans modifier le code.
