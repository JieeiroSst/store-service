FROM ollama/ollama:latest

WORKDIR /app

COPY vietnamese-vision-assistant.modelfile /app/vietnamese-vision-assistant.modelfile

RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

EXPOSE 11434

RUN echo '#!/bin/bash\n\
ollama serve &\n\
sleep 5\n\
ollama create vietnamese-vision-assistant -f /app/vietnamese-vision-assistant.modelfile\n\
tail -f /dev/null' > /app/start.sh && \
chmod +x /app/start.sh

ENTRYPOINT ["/app/start.sh"]