from copy import deepcopy
from datetime import datetime, timedelta

def _make_base_fs(fake_user: str, home_dir: str, fake_hostname: str) -> dict:
    return {
        "/": {"type": "dir", "children": ["home", "tmp", "etc", "var"]},
        "/home": {"type": "dir", "children": [fake_user]},
        f"{home_dir}": {"type": "dir", "children": ["notes.txt", "scripts", ".ssh"]},
        "/tmp": {"type": "dir", "children": []},
        "/etc": {"type": "dir", "children": ["passwd", "hostname", "shadow"]},
        "/var": {"type": "dir", "children": ["log"]},
        "/var/log": {"type": "dir", "children": ["auth.log"]},
        "/etc/hostname": {"type": "file", "content": f"{fake_hostname}\n"},
        "/etc/passwd": {"type": "file", "content": (
            "root:x:0:0:root:/root:/bin/bash\n"
            "daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin\n"
            "www-data:x:33:33:www-data:/var/www:/usr/sbin/nologin\n"
            f"{fake_user}:x:1000:1000:{fake_user}:{home_dir}:/bin/bash\n"
        )},
        "/etc/shadow": {"type": "file", "content": (
            "root:!:18930:0:99999:7:::\n"
            "daemon:*:18930:0:99999:7:::\n"
            "www-data:*:18930:0:99999:7:::\n"
            f"{fake_user}:$6$abcdef123456$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx:18930:0:99999:7:::\n"
        )},
        "/var/log/auth.log": {"type": "file", "content": (
            "Jan 22 10:12:01 honeypot sshd[1024]: Accepted password for user from 10.0.0.5 port 53422 ssh2\n"
            "Jan 22 10:12:03 honeypot sshd[1024]: pam_unix(sshd:session): session opened for user user(uid=1000)\n"
        )},
        f"{home_dir}/notes.txt": {"type": "file", "content": "TODO:\n- rotate ssh keys\n- backup /etc\n- check nginx logs\n"},
        f"{home_dir}/scripts": {"type": "dir", "children": ["backup.sh"]},
        f"{home_dir}/scripts/backup.sh": {"type": "file", "content": "#!/bin/bash\necho \"backup started\"\n"},
        f"{home_dir}/.ssh": {"type": "dir", "children": ["authorized_keys"]},
        f"{home_dir}/.ssh/authorized_keys": {"type": "file", "content": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCfakekey user@laptop\n"},
    }

def _copy_fs(base: dict) -> dict:
    return deepcopy(base)

class SessionState:
    def __init__(self, user: str, home_dir: str, hostname: str):
        self.user = user
        self.uid = 1000
        self.gid = 1000
        self.hostname = hostname
        self.home = home_dir
        self.cwd = home_dir
        self.fs = _copy_fs(_make_base_fs(user, home_dir, hostname))
        self.history = []
        self.time_offset = timedelta(seconds=0)
        self.fake_iface = "eth0"
        self.fake_ip = "192.168.1.10"
        self.fake_lo = "127.0.0.1"

    def now_local_str(self):
        dt = datetime.now().astimezone() + self.time_offset
        return dt.strftime("%a %b %e %H:%M:%S %Z %Y").replace("  ", " ")

    def add_history(self, cmd: str, max_n: int = 10):
        self.history.append(cmd)
        if len(self.history) > max_n:
            self.history = self.history[-max_n:]

def fs_exists(st: SessionState, path: str) -> bool:
    return path in st.fs

def fs_is_dir(st: SessionState, path: str) -> bool:
    node = st.fs.get(path)
    return bool(node) and node.get("type") == "dir"

def fs_list_dir(st: SessionState, path: str):
    node = st.fs.get(path)
    if not node or node.get("type") != "dir":
        return None
    return node.get("children", [])

def fs_read_file(st: SessionState, path: str):
    node = st.fs.get(path)
    if not node or node.get("type") != "file":
        return None
    return node.get("content", "")

def fs_write_file(st: SessionState, path: str, content: str):
    parent = "/" if path == "/" else "/".join(path.rstrip("/").split("/")[:-1])
    if parent == "":
        parent = "/"
    name = path.rstrip("/").split("/")[-1]
    if not fs_exists(st, parent) or not fs_is_dir(st, parent):
        return False, "No such file or directory"
    st.fs[path] = {"type": "file", "content": content}
    children = st.fs[parent].setdefault("children", [])
    if name not in children:
        children.append(name)
    return True, None

def fs_mkdir(st: SessionState, path: str):
    parent = "/" if path == "/" else "/".join(path.rstrip("/").split("/")[:-1])
    if parent == "":
        parent = "/"
    name = path.rstrip("/").split("/")[-1]
    if fs_exists(st, path):
        return False, "File exists"
    if not fs_exists(st, parent) or not fs_is_dir(st, parent):
        return False, "No such file or directory"
    st.fs[path] = {"type": "dir", "children": []}
    children = st.fs[parent].setdefault("children", [])
    if name not in children:
        children.append(name)
    return True, None

def fs_rm(st: SessionState, path: str):
    if not fs_exists(st, path):
        return False, "No such file or directory"
    if fs_is_dir(st, path):
        children = fs_list_dir(st, path) or []
        if children:
            return False, "Is a directory"
    parent = "/" if path == "/" else "/".join(path.rstrip("/").split("/")[:-1])
    if parent == "":
        parent = "/"
    name = path.rstrip("/").split("/")[-1]
    st.fs.pop(path, None)
    if fs_exists(st, parent) and fs_is_dir(st, parent):
        kids = st.fs[parent].setdefault("children", [])
        if name in kids:
            kids.remove(name)
    return True, None

def norm_path(st: SessionState, path: str) -> str:
    if path is None:
        return st.cwd
    path = path.strip()
    if path == "" or path == ".":
        return st.cwd
    if path == "~" or path.startswith("~/"):
        path = st.home + path[1:]
    if path.startswith("/"):
        p = path
    else:
        base = "" if st.cwd == "/" else st.cwd
        p = f"{base}/{path}"
    parts = []
    for part in p.split("/"):
        if part in ("", "."):
            continue
        if part == "..":
            if parts:
                parts.pop()
            continue
        parts.append(part)
    return "/" if not parts else "/" + "/".join(parts)

def fs_snapshot(st: SessionState, max_lines: int = 20) -> str:
    paths = ["/", "/home", st.home, f"{st.home}/scripts", f"{st.home}/.ssh", "/etc", "/var/log", "/tmp", "/etc/hostname", "/etc/passwd", "/etc/shadow", "/var/log/auth.log", f"{st.home}/notes.txt", f"{st.home}/scripts/backup.sh", f"{st.home}/.ssh/authorized_keys"]
    lines = []
    for p in paths:
        if fs_exists(st, p):
            node = st.fs[p]
            if node.get("type") == "dir":
                kids = node.get("children", [])
                lines.append(f"{p}/ (dir): " + ", ".join(kids[:20]))
            else:
                content = node.get("content", "")
                preview = content.replace("\n", "\\n")
                if len(preview) > 120:
                    preview = preview[:120] + "..."
                lines.append(f"{p} (file): {preview}")
        if len(lines) >= max_lines:
            break
    return "\n".join(lines)
