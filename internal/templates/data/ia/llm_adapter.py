import requests
from fs_engine import fs_snapshot
from utils import log_event

def build_shell_prompt(st, cmd: str) -> str:
    # On donne juste le strict minimum + des exemples pour forcer le format
    snap = fs_snapshot(st)
    return f"""You are a Linux Shell. Behave exactly like a terminal.
RULES:
1. Do NOT explain. Do NOT chat.
2. Output ONLY the standard stdout/stderr.
3. If the command is silent (like 'cd', 'mkdir', 'export'), output nothing.
4. If the command is not found, output 'bash: {cmd}: command not found'.

CONTEXT:
User: {st.user}
Dir: {st.cwd}
Files: {snap}

EXAMPLES:
Cmd: whoami
Out: {st.user}

Cmd: cd /tmp
Out:

Cmd: pwd
Out: /tmp

Cmd: notarealcommand
Out: bash: notarealcommand: command not found

CURRENT COMMAND:
Cmd: {cmd}
Out:"""
def post_validate_output(st, cmd: str, output: str) -> str:
    c = cmd.strip()
    if c == "whoami":
        return st.user + "\n"
    if c == "pwd":
        return st.cwd + "\n"
    if c == "hostname":
        return st.hostname + "\n"
    if c == "id":
        return f"uid={st.uid}({st.user}) gid={st.gid}({st.user}) groups={st.gid}({st.user})\n"
    if c == "date":
        return st.now_local_str() + "\n"
    cleaned = output.replace("```", "").strip("\n")
    return cleaned + ("\n" if not cleaned.endswith("\n") else "")

def ollama_shell_reply(st, cmd: str, session: requests.Session, session_id: str, addr, ollama_url: str, ollama_model: str) -> tuple[str, int]:
    prompt = build_shell_prompt(st, cmd)
    payload = {
        "model": ollama_model,
        "prompt": prompt,
        "stream": False,
        "options": {
            "temperature": 0.3,
            "num_predict": 200,
            "top_k": 20,
            "top_p": 0.85,
            "repeat_penalty": 1.05,
        }
    }
    try:
        r = session.post(ollama_url, json=payload, timeout=35)
        r.raise_for_status()
        data = r.json()
        response = (data.get("response") or "").strip()
        if response:
            out = post_validate_output(st, cmd, response)
            log_event(session_id, addr, "llm_success", {"cmd": cmd, "response_preview": out[:200]})
            return out, 0
        log_event(session_id, addr, "llm_empty", {"cmd": cmd})
        return (f"bash: {cmd}: command not found\n", 127)
    except requests.exceptions.Timeout:
        log_event(session_id, addr, "llm_timeout", {"cmd": cmd})
        # deterministic fallback when LLM times out
        return (f"bash: {cmd}: LLM unavailable (timeout)\n", 127)
    except requests.exceptions.ConnectionError as e:
        log_event(session_id, addr, "llm_connection_error", {"cmd": cmd, "error": str(e)})
        # deterministic fallback when LLM can't be reached
        return (f"bash: {cmd}: LLM unavailable (connection)\n", 127)
    except Exception as e:
        log_event(session_id, addr, "llm_error", {"cmd": cmd, "error": type(e).__name__, "details": str(e)})
        # generic deterministic fallback
        return (f"bash: {cmd}: LLM error ({type(e).__name__})\n", 127)
