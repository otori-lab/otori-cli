# Démo : SSH Honeypot

Exemples de commandes et comportements attendus

1) Connexion
```bash
ssh -p 2222 user@localhost
# mot de passe : password
```

2) Commandes de base
- `whoami` -> `user`
- `pwd` -> `/home/user`
- `hostname` -> `honeypot`
- `date` -> date simulée

3) Fichiers et FS en mémoire (persistés après session)
```bash
cat /home/user/notes.txt
echo "hello" > test.txt
cat test.txt
echo "more" >> test.txt
cp test.txt copy.txt
mv copy.txt moved.txt
ls
rm test.txt
```

4) Où trouver les sessions
- Événements (JSONL) : `logs/honeypot_sessions.jsonl`
- État complet session (FS + history) : `logs/sessions/<session_id>.json`

5) Comportement LLM
- Si Ollama est disponible, certaines commandes non triviales sont passées à l'IA (avec garde-fous).
- Si Ollama est indisponible, le serveur renverra une réponse déterministe et sûre (pas de modification FS faite par l'IA).

6) Notes
- Chaque session a un FS indépendant en mémoire ; les mutations sont visibles durant la session et sauvegardées pour la démo.
