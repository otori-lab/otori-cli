import socket
import threading
from datetime import datetime, timezone, timedelta
from pathlib import Path
import json
import uuid
import requests
import paramiko
import shlex
import re
import time

# Externalized helpers (will override local implementations)
from utils import utc_now, log_event, to_crlf
from fs_engine import (
    SessionState,
    fs_exists,
    fs_is_dir,
    fs_list_dir,
    fs_read_file,
    fs_write_file,
    fs_mkdir,
    fs_rm,
    norm_path,
    fs_snapshot,
)
from llm_adapter import ollama_shell_reply

# =========================
# CONFIG
# =========================
LISTEN_HOST = "0.0.0.0"
LISTEN_PORT = 2222

FAKE_USER = "user"
FAKE_PASS = "password"

HOSTKEY_PATH = Path("hostkey_rsa")
FAKE_HOSTNAME = "honeypot"

LOG_DIR = Path("logs")
HOME_DIR = f"/home/{FAKE_USER}"

# =========================
# BLACKLIST & HYBRID MODE
# =========================
# Commandes qui sont trop longues/complexes pour répondre directement
COMMAND_BLACKLIST = [
    "curl", "wget", "nc", "ncat", "socat", "telnet",  # Network tools
    "sed", "awk", "perl", "python", "ruby", "bash",   # Interpreters/scripting
    "find", "locate", "updatedb",                      # Search tools
    "tar", "gzip", "bzip2", "zip", "unzip",           # Compression
    "gcc", "make", "cmake", "g++",                     # Compilation
    "docker", "kubectl", "systemctl", "journalctl",   # System admin
    "ssh", "scp", "rsync",                             # Remote tools
    "gcc", "nm", "objdump", "strings",                # Binary analysis
]



# =========================
# OLLAMA CONFIG
# =========================
OLLAMA_URL = "http://127.0.0.1:11434/api/generate"
OLLAMA_MODEL = "mistral:latest"

# =========================
# SSH KEY
# =========================
def load_or_create_hostkey():
    if HOSTKEY_PATH.exists():
        return paramiko.RSAKey(filename=str(HOSTKEY_PATH))
    key = paramiko.RSAKey.generate(2048)
    key.write_private_key_file(str(HOSTKEY_PATH))
    return key

# =========================
# PARAMIKO SERVER
# =========================
class HoneypotServer(paramiko.ServerInterface):
    def __init__(self, client_addr):
        self.client_addr = client_addr
        self.shell_event = threading.Event()

    def check_auth_password(self, username, password):
        print(f"[{utc_now()}] Auth attempt from {self.client_addr} user={username} pass={password}")
        if username == FAKE_USER and password == FAKE_PASS:
            return paramiko.AUTH_SUCCESSFUL
        return paramiko.AUTH_FAILED

    def get_allowed_auths(self, username):
        return "password"

    def check_channel_request(self, kind, chanid):
        if kind == "session":
            return paramiko.OPEN_SUCCEEDED
        return paramiko.OPEN_FAILED_ADMINISTRATIVELY_PROHIBITED

    def check_channel_pty_request(self, channel, term, width, height, pixelwidth, pixelheight, modes):
        return True

    def check_channel_shell_request(self, channel):
        self.shell_event.set()
        return True

# =========================
# READ LINE (SSH INPUT)
# =========================
def read_line(chan):
    buf = ""
    while True:
        data = chan.recv(1)
        if not data:
            return None
        ch = data.decode("utf-8", errors="ignore")

        if ch in ("\r", "\n"):
            try:
                chan.send("\r\n")
            except Exception:
                pass
            return buf

        if ch in ("\x7f", "\x08"):
            if buf:
                buf = buf[:-1]
                try:
                    chan.send("\b \b")
                except Exception:
                    pass
            continue

        if ord(ch) < 32 and ch not in ("\t",):
            continue

        try:
            chan.send(ch)
        except Exception:
            return None

        buf += ch


def quick_command(st: SessionState, cmd: str) -> tuple[str | None, int | None]:
    """Handle trivial, stateless commands that don't need FS context."""
    c = cmd.strip()
    if c == "whoami":
        return st.user + "\n", 0
    if c == "hostname":
        return st.hostname + "\n", 0
    if c == "pwd":
        return st.cwd + "\n", 0
    if c == "id":
        return f"uid={st.uid}({st.user}) gid={st.gid}({st.user}) groups={st.gid}({st.user})\n", 0
    if c == "date":
        return st.now_local_str() + "\n", 0
    if c in ("clear", "reset"):
        # We can't actually clear the client terminal, but we can send the escape code
        return "\x1b[H\x1b[2J", 0
    return None, None


def handle_fs_ops(st: SessionState, cmd: str) -> tuple[str | None, int | None]:
    """Handle commands that read or mutate the virtual filesystem."""
    try:
        parts = shlex.split(cmd)
    except ValueError:
        return "bash: unclosed quote\n", 1

    if not parts:
        return "", 0

    command = parts[0]
    args = parts[1:]

    if command == "cd":
        path = args[0] if args else "~"
        new_path = norm_path(st, path)
        if not fs_exists(st, new_path):
            return f"bash: cd: {path}: No such file or directory\n", 1
        if not fs_is_dir(st, new_path):
            return f"bash: cd: {path}: Not a directory\n", 1
        st.cwd = new_path
        return "", 0

    if command == "ls":
        path = args[0] if args else "."
        target_path = norm_path(st, path)
        if not fs_exists(st, target_path) or not fs_is_dir(st, target_path):
            return f"ls: cannot access '{path}': No such file or directory\n", 2
        
        children = fs_list_dir(st, target_path)
        if children is None:
             return f"ls: cannot access '{path}': Not a directory\n", 2

        return "  ".join(sorted(children)) + "\n", 0

    if command == "cat":
        if not args:
            return "", 0  # cat without args waits for stdin, we just return nothing
        path = args[0]
        target_path = norm_path(st, path)
        
        if fs_exists(st, target_path):
            if fs_is_dir(st, target_path):
                return f"cat: {path}: Is a directory\n", 1
            content = fs_read_file(st, target_path)
            return content + "\n" if not content.endswith("\n") else content, 0
        else:
            # file does not exist, let LLM handle it
            return None, None


    if command == "mkdir":
        if not args:
            return "mkdir: missing operand\n", 1
        path = args[0]
        target_path = norm_path(st, path)
        success, err = fs_mkdir(st, target_path)
        if not success:
            return f"mkdir: cannot create directory ‘{path}’: {err}\n", 1
        return "", 0

    if command == "touch":
        if not args:
            return "touch: missing file operand\n", 1
        path = args[0]
        target_path = norm_path(st, path)
        if not fs_exists(st, target_path):
             # create empty file
             success, err = fs_write_file(st, target_path, "")
             if not success:
                 return f"touch: cannot touch '{path}': {err}\n", 1
        else:
            # TODO: update timestamp if file exists
            pass
        return "", 0

    if command == "rm":
        if not args:
            return "rm: missing operand\n", 1
        path = args[0]
        target_path = norm_path(st, path)
        success, err = fs_rm(st, target_path)
        if not success:
            return f"rm: cannot remove '{path}': {err}\n", 1
        return "", 0

    return None, None


# =========================
# HANDLE CLIENT (CORE)
# =========================
def handle_client(client_sock, addr, host_key):
    transport = paramiko.Transport(client_sock)
    transport.add_server_key(host_key)

    session_id = str(uuid.uuid4())
    log_event(session_id, addr, "session_start", {})

    server = HoneypotServer(addr)
    try:
        transport.start_server(server=server)
    except paramiko.SSHException as e:
        log_event(session_id, addr, "ssh_negotiation_failed", {"error": repr(e)})
        try:
            transport.close()
        except Exception:
            pass
        return

    chan = transport.accept(20)
    if chan is None:
        log_event(session_id, addr, "no_channel", {})
        try:
            transport.close()
        except Exception:
            pass
        return

    server.shell_event.wait(10)
    if not server.shell_event.is_set():
        log_event(session_id, addr, "no_shell_request", {})
        try:
            chan.close()
        except Exception:
            pass
        try:
            transport.close()
        except Exception:
            pass
        return

    st = SessionState(user=FAKE_USER, home_dir=HOME_DIR, hostname=FAKE_HOSTNAME)
    http = requests.Session()

    def prompt():
        return f"{st.user}@{st.hostname}:{st.cwd}$ "

    # banner
    try:
        chan.send("Welcome to Ubuntu 22.04.4 LTS (GNU/Linux 5.15.0-xx-generic x86_64)\r\n")
        chan.send(f"Last login: {utc_now()}\r\n\r\n")
    except Exception as e:
        log_event(session_id, addr, "send_failed", {"stage": "banner", "error": repr(e)})
        try:
            chan.close()
        except Exception:
            pass
        try:
            transport.close()
        except Exception:
            pass
        return

    while True:
        try:
            if chan.closed or not transport.is_active():
                break
            chan.send(prompt())
        except Exception as e:
            log_event(session_id, addr, "send_failed", {"stage": "prompt", "error": repr(e)})
            break

        try:
            cmd = read_line(chan)
        except Exception as e:
            log_event(session_id, addr, "read_failed", {"error": repr(e)})
            break

        if cmd is None:
            break

        cmd = cmd.strip()
        if cmd == "":
            continue

        st.add_history(cmd)
        log_event(session_id, addr, "command", {"cmd": cmd, "cwd": st.cwd})

        if cmd in ("exit", "logout"):
            output = "logout\r\n"
            try:
                if not chan.closed and transport.is_active():
                    chan.send(output)
            except Exception:
                pass
            log_event(session_id, addr, "output", {"output": output, "exit_code": 0})
            break

        exit_code = 0
        output = ""

        qout, qcode = quick_command(st, cmd)
        
        if qout is not None:
            output, exit_code = qout, qcode
        else:
            fout, fcode = handle_fs_ops(st, cmd)
            if fout is not None:
                output, exit_code = fout, fcode
            else:
                # 3) everything else -> LLM (but state-aware + guardrails)
                is_cat_on_nonexistent = False
                cat_path = None
                try:
                    parts = shlex.split(cmd)
                    if parts and parts[0] == 'cat' and len(parts) > 1:
                        path = parts[1]
                        target_path = norm_path(st, path)
                        if not fs_exists(st, target_path):
                            is_cat_on_nonexistent = True
                            cat_path = target_path
                except ValueError:
                    pass

                print(f"[DEBUG] Using LLM for: '{cmd}'")
                out, code = ollama_shell_reply(st, cmd, http, session_id, addr, OLLAMA_URL, OLLAMA_MODEL)
                
                if is_cat_on_nonexistent and out and code == 0:
                    # The LLM returned content for a file that didn't exist. Let's create it.
                    print(f"[DEBUG] LLM generated content for non-existent file '{cat_path}'. Creating it.")
                    fs_write_file(st, cat_path, out.strip())

                output, exit_code = out, code
                print(f"[DEBUG] LLM result: {output[:50]}")

        # send output safely (CRLF)
        try:
            if chan.closed or not transport.is_active():
                break
            chan.send(to_crlf(output))
        except Exception as e:
            log_event(session_id, addr, "send_failed", {"stage": "output", "error": repr(e)})
            break

        log_event(session_id, addr, "output", {"output": output[:1000], "exit_code": exit_code})

    log_event(session_id, addr, "session_end", {})

    # Persist session state for demo / analysis
    try:
        sess_dir = LOG_DIR / "sessions"
        sess_dir.mkdir(parents=True, exist_ok=True)
        sess_file = sess_dir / f"{session_id}.json"
        with sess_file.open("w", encoding="utf-8") as f:
            json.dump({
                "ts": utc_now(),
                "sid": session_id,
                "user": st.user,
                "hostname": st.hostname,
                "cwd": st.cwd,
                "history": st.history,
                "fs": st.fs,
            }, f, ensure_ascii=False, indent=2)
        log_event(session_id, addr, "session_saved", {"path": str(sess_file)})
    except Exception as e:
        log_event(session_id, addr, "session_save_failed", {"error": repr(e)})

    try:
        chan.close()
    except Exception:
        pass
    try:
        transport.close()
    except Exception:
        pass
    try:
        http.close()
    except Exception:
        pass

# =========================
# MAIN
# =========================
def main():
    host_key = load_or_create_hostkey()

    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.bind((LISTEN_HOST, LISTEN_PORT))
    sock.listen(100)

    print(f"[+] SSH honeypot listening on {LISTEN_HOST}:{LISTEN_PORT}")
    print(f"[i] Login: {FAKE_USER} / {FAKE_PASS}")
    print(f"[i] Hostkey saved at: {HOSTKEY_PATH.resolve()}")
    print(f"[i] Ollama: {OLLAMA_URL}")
    print(f"[i] model={OLLAMA_MODEL}")

    while True:
        client, addr = sock.accept()
        t = threading.Thread(target=handle_client, args=(client, addr, host_key), daemon=True)
        t.start()

if __name__ == "__main__":
    main()
