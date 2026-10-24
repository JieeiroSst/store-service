FROM ollama/ollama:0.40.10

RUN useradd -u 10001 -m -d /home/ollama ollama-runner

ENV HOME=/home/ollama \
    OLLAMA_MODELS=/home/ollama/.ollama/models \
    OLLAMA_HOST=0.0.0.0:11434

WORKDIR /app

COPY vietnamese-vision-assistant.modelfile /app/vietnamese-vision-assistant.modelfile

RUN printf '%s\n' \
    '#!/bin/sh' \
    'set -e' \
    'ollama serve &' \
    'server=$!' \
    'until ollama list >/dev/null 2>&1; do' \
    '  kill -0 "$server" 2>/dev/null || exit 1' \
    '  sleep 1' \
    'done' \
    'ollama create vietnamese-vision-assistant -f /app/vietnamese-vision-assistant.modelfile' \
    'wait "$server"' \
    > /app/start.sh && chmod 0755 /app/start.sh

EXPOSE 11434

USER 10001

ENTRYPOINT ["/app/start.sh"]
