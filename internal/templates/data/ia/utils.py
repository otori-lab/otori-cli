from pathlib import Path
import json
from datetime import datetime, timezone

LOG_DIR = Path("logs")

def utc_now() -> str:
    return datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")

def log_event(session_id: str, client_addr, event_type: str, data: dict):
    """
    Log event in simplified JSONL format.
    """
    LOG_DIR.mkdir(exist_ok=True)
    
    entry = {
        "ts": utc_now(),
        "sid": session_id,
        "ip": client_addr[0],
        "port": client_addr[1],
        "type": event_type,
    }
    
    # Add relevant fields based on event type
    if event_type == "command":
        entry["cmd"] = data.get("cmd", "")
        entry["cwd"] = data.get("cwd", "")
    elif event_type == "output":
        entry["cmd"] = data.get("cmd", "")
        entry["out"] = data.get("output", "")[:500]  # limit output length
        entry["code"] = data.get("exit_code", 0)
    else:
        entry.update(data or {})

    with (LOG_DIR / "honeypot_sessions.jsonl").open("a", encoding="utf-8") as f:
        f.write(json.dumps(entry, ensure_ascii=False) + "\n")

def to_crlf(s: str) -> str:
    if s is None:
        return ""
    s = s.replace("\r\n", "\n").replace("\r", "\n")
    return s.replace("\n", "\r\n")
