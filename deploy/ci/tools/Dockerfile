# Utility image for the in-cluster CI pipeline (deploy/ci): git for the repo
# mirror, python for change detection, kubectl for pipeline state, docker CLI
# for builds against the node's Docker daemon.
# The pipeline scripts themselves come from a ConfigMap (deploy/ci/scripts).
FROM docker:29-cli

ARG KUBECTL_VERSION=v1.36.1
RUN apk add --no-cache git python3 py3-yaml bash curl \
 && arch="$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')" \
 && curl -fsSL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${arch}/kubectl" \
 && chmod +x /usr/local/bin/kubectl \
 && git config --system --add safe.directory '*'

