# LLM SSH Honeypot

Honeypot SSH avec reponses generees par IA (Ollama).

## Demarrage

```bash
docker compose up -d
```

Le modele Mistral sera telecharge automatiquement au premier usage.

## Test

```bash
ssh -p PORT admin@localhost
# Password: admin123
```

## Logs

Les logs sont dans `./logs/honeypot_sessions.jsonl`

## Configuration

Modifiez les variables d'environnement dans `docker-compose.yml`:
- `FAKE_USER` / `FAKE_PASS` : credentials acceptes
- `FAKE_HOSTNAME` : hostname affiche
- `EXTRA_CONTEXT` : contexte additionnel pour le LLM
