FROM python:3.11-slim

# Install system dependencies (curl is required for installing Ollama)
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Install Ollama
RUN curl -fsSL https://ollama.com/install.sh | sh

# Set working directory
WORKDIR /app

# Copy application code
COPY . /app

# Install Python dependencies
RUN pip install --no-cache-dir requests paramiko

# Create a startup script to run Ollama and the Honeypot together
RUN echo '#!/bin/bash' > /app/start.sh && \
    echo 'ollama serve &' >> /app/start.sh && \
    echo 'sleep 5' >> /app/start.sh && \
    echo 'echo "Checking/Pulling Mistral model..."' >> /app/start.sh && \
    echo 'ollama pull mistral' >> /app/start.sh && \
    echo 'echo "Starting Honeypot..."' >> /app/start.sh && \
    echo 'python honeypot_ssh.py' >> /app/start.sh && \
    chmod +x /app/start.sh

# Expose SSH port
EXPOSE 2222

# Start everything
CMD ["/app/start.sh"]