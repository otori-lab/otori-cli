FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY log_shipper.py .

ENV MONITORING_URL=""
ENV PROFILE_NAME=""
ENV FAKE_HOSTNAME="cowrie"
ENV LOG_FILE="/logs/cowrie.json"
ENV DATA_DIR="/data"

CMD ["python", "-u", "log_shipper.py"]
