# Otori CLI

Otori CLI est un outil en ligne de commande permettant de déployer et d’expérimenter des honeypots locaux à des fins pédagogiques, expérimentales et de recherche en cybersécurité.

Projet développé dans le cadre d’un Projet de Fin d’Études (PFE). L’outil est conçu pour être léger, local et open source.

---

## Commandes

```bash
otori init
otori deploy
otori status
otori stop
````

---

## otori init

Initialise un profil de honeypot.

### Mode interactif

```bash
otori init
```

### Mode non interactif

```bash
otori init \
  --type [classic|ia] \
  --profile-name [profile_name] \
  --server-name [server_name] \
  --company [company_name] \
  --users root,admin,test
```

### Flags

* --type *(obligatoire)* : type de honeypot
* --profile-name *(optionnel)* : nom du profil (défaut : default)
* --server-name *(obligatoire)* : nom du serveur simulé
* --company *(optionnel)* : organisation simulée
* --users *(optionnel)* : utilisateurs fictifs

---

## otori deploy

Déploie un honeypot à partir d’un profil.

```bash
otori deploy
otori deploy --profile srv04
```

### Flags

* --profile : profil à utiliser
* --force : force le redéploiement

---

## otori status

Affiche l’état du honeypot local.

```bash
otori status
otori status --profile srv04
otori status --json
```

### Flags

* --profile : statut d’un profil spécifique
* --json : sortie JSON

---

## otori stop

Arrête le honeypot en cours.

```bash
otori stop
otori stop --force
```

### Flags

* --force : force l’arrêt

---

## Statut

Projet en cours de développement.

---

## Licence

MIT