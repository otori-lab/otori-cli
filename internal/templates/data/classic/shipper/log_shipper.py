"""
Cowrie Log Shipper - Sidecar that reads Cowrie logs and sends them to monitoring.

Features:
- Automatic registration with monitoring server on startup
- Tail of Cowrie JSON log file with real-time shipping
- Disk buffer for robustness (retry on failure)
- Conversion from Cowrie format to OtoriEventIn
"""

import json
import logging
import os
import signal
import socket
import sys
import threading
import time
from pathlib import Path

import requests

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)
logger = logging.getLogger(__name__)

# =========================
# CONFIG
# =========================
MONITORING_URL = os.environ.get("MONITORING_URL", "")
PROFILE_NAME = os.environ.get("PROFILE_NAME", "")
FAKE_HOSTNAME = os.environ.get("FAKE_HOSTNAME", "cowrie")

DATA_DIR = Path(os.environ.get("DATA_DIR", "/data"))
LOG_FILE = Path(os.environ.get("LOG_FILE", "/logs/cowrie.json"))

SENSOR_FILE = DATA_DIR / "sensor.json"
POSITION_FILE = DATA_DIR / "position.txt"
BUFFER_FILE = DATA_DIR / "buffer.jsonl"

RETRY_INTERVAL = 30  # seconds
SHIP_INTERVAL = 1  # seconds
REGISTRATION_RETRY_INTERVAL = 60  # seconds


def _get_local_ip() -> str:
    """Get the local IP address of the machine."""
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except Exception:
        return "127.0.0.1"


class CowrieLogShipper:
    """Sends Cowrie honeypot logs to the monitoring server."""

    def __init__(self):
        self.monitoring_url = MONITORING_URL.rstrip("/")
        self.sensor_id = None
        self.token = None
        self.running = False
        self._http = requests.Session()
        self._ship_thread = None
        self._retry_thread = None

    def is_enabled(self) -> bool:
        """Check if monitoring is configured."""
        return bool(self.monitoring_url)

    def register(self) -> bool:
        """Register the honeypot with the monitoring server."""
        if not self.is_enabled():
            logger.info("Monitoring disabled (MONITORING_URL not set)")
            return False

        # Check if already registered
        if SENSOR_FILE.exists():
            try:
                data = json.loads(SENSOR_FILE.read_text())
                self.sensor_id = data["sensor_id"]
                self.token = data.get("token")
                logger.info(f"Sensor already registered: {self.sensor_id}")
                return True
            except Exception as e:
                logger.warning(f"Failed to read sensor file: {e}")

        # Registration
        try:
            resp = self._http.post(
                f"{self.monitoring_url}/register",
                json={
                    "hostname": FAKE_HOSTNAME,
                    "honeypot_type": "classic",
                    "ip": _get_local_ip(),
                    "profile_name": PROFILE_NAME or None,
                },
                timeout=10,
            )
            resp.raise_for_status()
            data = resp.json()
            self.sensor_id = data["sensor_id"]
            self.token = data.get("token")

            # Save for next startups
            DATA_DIR.mkdir(parents=True, exist_ok=True)
            SENSOR_FILE.write_text(json.dumps(data, indent=2))

            logger.info(f"Sensor registered: {self.sensor_id}")
            return True

        except requests.exceptions.ConnectionError:
            logger.error(f"Cannot connect to monitoring: {self.monitoring_url}")
            return False
        except requests.exceptions.Timeout:
            logger.error("Registration timeout")
            return False
        except Exception as e:
            logger.error(f"Registration failed: {e}")
            return False

    def _convert_event(self, raw: dict) -> dict | None:
        """Convert a Cowrie event to OtoriEventIn format."""
        eventid = raw.get("eventid", "")

        # Mapping Cowrie eventid -> event_type
        event_map = {
            "cowrie.session.connect": "connect",
            "cowrie.login.success": "login_success",
            "cowrie.login.failed": "login_failed",
            "cowrie.command.input": "command",
            "cowrie.session.closed": "closed",
        }

        event_type = event_map.get(eventid)
        if not event_type:
            return None  # Ignore other events

        return {
            "timestamp": raw.get("timestamp", ""),
            "sensor": self.sensor_id,
            "honeypot_type": "classic",
            "session_id": raw.get("session"),
            "src_ip": raw.get("src_ip"),
            "src_port": raw.get("src_port"),
            "dst_ip": raw.get("dst_ip"),
            "dst_port": raw.get("dst_port"),
            "event_type": event_type,
            "command": raw.get("input"),  # Cowrie uses "input" not "cmd"
            "username": raw.get("username"),
            "password": raw.get("password"),
            "duration_sec": raw.get("duration"),
        }

    def _ship_event(self, event: dict) -> bool:
        """Send an event to the monitoring server."""
        try:
            resp = self._http.post(
                f"{self.monitoring_url}/ingest",
                json=event,
                timeout=5,
            )
            return resp.status_code == 200
        except Exception as e:
            logger.debug(f"Ship failed: {e}")
            return False

    def _buffer_event(self, event: dict):
        """Write an event to the disk buffer."""
        try:
            DATA_DIR.mkdir(parents=True, exist_ok=True)
            with BUFFER_FILE.open("a", encoding="utf-8") as f:
                f.write(json.dumps(event) + "\n")
        except Exception as e:
            logger.error(f"Buffer write failed: {e}")

    def _get_position(self) -> int:
        """Get the saved read position."""
        try:
            if POSITION_FILE.exists():
                return int(POSITION_FILE.read_text().strip())
        except Exception:
            pass
        return 0

    def _save_position(self, pos: int):
        """Save the read position."""
        try:
            DATA_DIR.mkdir(parents=True, exist_ok=True)
            POSITION_FILE.write_text(str(pos))
        except Exception as e:
            logger.error(f"Position save failed: {e}")

    def _ship_loop(self):
        """Main thread: tail the log file and ship events."""
        position = self._get_position()
        logger.info(f"Starting ship loop from position {position}")

        while self.running:
            try:
                if not LOG_FILE.exists():
                    logger.debug(f"Waiting for log file: {LOG_FILE}")
                    time.sleep(SHIP_INTERVAL)
                    continue

                with LOG_FILE.open("r", encoding="utf-8") as f:
                    f.seek(position)

                    for line in f:
                        if not self.running:
                            break

                        line = line.strip()
                        if not line:
                            continue

                        try:
                            raw = json.loads(line)
                        except json.JSONDecodeError:
                            continue

                        event = self._convert_event(raw)
                        if not event:
                            continue

                        if self._ship_event(event):
                            sid = event.get("session_id", "")[:8] if event.get("session_id") else "?"
                            logger.debug(f"Shipped: {event.get('event_type')} sid={sid}")
                        else:
                            self._buffer_event(event)
                            logger.warning(f"Buffered: {event.get('event_type')}")

                    position = f.tell()
                    self._save_position(position)

            except Exception as e:
                logger.error(f"Ship loop error: {e}")

            time.sleep(SHIP_INTERVAL)

    def _retry_loop(self):
        """Retry thread: reread buffer and retry failed events."""
        while self.running:
            time.sleep(RETRY_INTERVAL)

            if not BUFFER_FILE.exists():
                continue

            try:
                # Read buffer
                with BUFFER_FILE.open("r", encoding="utf-8") as f:
                    lines = f.readlines()

                if not lines:
                    continue

                logger.info(f"Retrying {len(lines)} buffered events...")

                # Retry each event
                failed = []
                for line in lines:
                    line = line.strip()
                    if not line:
                        continue

                    try:
                        event = json.loads(line)
                    except json.JSONDecodeError:
                        continue

                    if not self._ship_event(event):
                        failed.append(line)

                # Rewrite failed events
                if failed:
                    with BUFFER_FILE.open("w", encoding="utf-8") as f:
                        f.write("\n".join(failed) + "\n")
                    logger.info(f"Retry: {len(lines) - len(failed)} success, {len(failed)} still buffered")
                else:
                    BUFFER_FILE.unlink(missing_ok=True)
                    logger.info(f"Retry: all {len(lines)} events shipped successfully")

            except Exception as e:
                logger.error(f"Retry loop error: {e}")

    def start(self):
        """Start the shipping threads."""
        if not self.is_enabled():
            logger.info("Monitoring not configured, shipper will not start")
            return False

        # Try to register, retry if fails
        while not self.sensor_id:
            if self.register():
                break
            logger.warning(f"Registration failed, retrying in {REGISTRATION_RETRY_INTERVAL}s...")
            time.sleep(REGISTRATION_RETRY_INTERVAL)

        self.running = True

        self._ship_thread = threading.Thread(target=self._ship_loop, daemon=True, name="log-shipper")
        self._ship_thread.start()

        self._retry_thread = threading.Thread(target=self._retry_loop, daemon=True, name="log-retry")
        self._retry_thread.start()

        logger.info("Cowrie log shipper started")
        return True

    def stop(self):
        """Stop the threads."""
        self.running = False
        logger.info("Cowrie log shipper stopped")


def main():
    """Main entry point."""
    logger.info("=" * 50)
    logger.info("Cowrie Log Shipper - Sidecar")
    logger.info("=" * 50)
    logger.info(f"MONITORING_URL: {MONITORING_URL or '(not set)'}")
    logger.info(f"PROFILE_NAME: {PROFILE_NAME or '(not set)'}")
    logger.info(f"FAKE_HOSTNAME: {FAKE_HOSTNAME}")
    logger.info(f"LOG_FILE: {LOG_FILE}")
    logger.info(f"DATA_DIR: {DATA_DIR}")
    logger.info("=" * 50)

    shipper = CowrieLogShipper()

    # Handle graceful shutdown
    def signal_handler(sig, frame):
        logger.info("Shutdown signal received")
        shipper.stop()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    if not shipper.is_enabled():
        logger.warning("MONITORING_URL not set - shipper inactive, waiting for configuration...")
        # Keep running but do nothing (allows container to stay up for debugging)
        while True:
            time.sleep(60)

    if shipper.start():
        # Keep main thread alive
        while shipper.running:
            time.sleep(1)
    else:
        logger.error("Failed to start shipper")
        sys.exit(1)


if __name__ == "__main__":
    main()
