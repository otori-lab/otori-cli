# SSH Honeypot

Instructions pour exécuter le honeypot SSH sur Windows (PowerShell)

Prérequis
- Python 3.9+
- (Optionnel) Ollama local si vous souhaitez les réponses LLM (voir `OLLAMA_URL` dans `honeypot_ssh.py`).

Étapes (PowerShell)

```powershell
python -m venv .venv
.\.venv\Scripts\Activate.ps1
pip install -r requirements.txt
python honeypot_ssh.py
```

Notes
- Le serveur écoute sur le port 2222 (non privilégié). Connectez-vous avec : `ssh -p 2222 user@HOST` et le mot de passe `password`.
- La clé d'hôte est `hostkey_rsa` (déjà présente) — le script la créera si elle manque.
- Les logs sont écrits dans le dossier `logs/honeypot_sessions.jsonl`.
- Si vous n'avez pas un service Ollama local sur `http://127.0.0.1:11434`, les commandes non reconnues retourneront une erreur générique (le honeypot fonctionne sans Ollama).

Pour une démo complète et exemples de commandes, voir `DEMO.md`.
