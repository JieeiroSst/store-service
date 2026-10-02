#!/usr/bin/env bash
# Local GitOps bootstrap for a docker-desktop cluster.
#
#   deploy/bootstrap-local.sh install              # Argo CD + add-ons + local CI + store-dev
#   deploy/bootstrap-local.sh remove-default-apps  # drop app workloads from `default`
#   deploy/bootstrap-local.sh ci-run [all|svc,...] # build services now (default: what changed)
#   deploy/bootstrap-local.sh status
#
# After `install`, the loop is fully automatic and never touches GitHub Actions:
#   git push master -> the in-cluster pipeline (deploy/ci, Argo Workflows)
#   notices the commit within a minute, tests and builds the changed services
#   into the local registry (localhost:5000) -> Argo CD Image Updater sets the
#   new tag on <service>-dev -> Argo CD rolls store-dev.
set -euo pipefail

CONTEXT="docker-desktop"
ARGOCD_CHART_VERSION="10.9.6"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ARGO_DIR="$ROOT/deploy/argocd"
UMBRELLA_RELEASE="chart-1785009827"

kc() { kubectl --context "$CONTEXT" "$@"; }

require_context() {
  if ! kubectl config get-contexts -o name | grep -qx "$CONTEXT"; then
    echo "kube context $CONTEXT not found" >&2
    exit 1
  fi
}

install() {
  require_context
  helm repo add argo https://argoproj.github.io/argo-helm >/dev/null 2>&1 || true
  helm repo update argo >/dev/null

  echo "==> Argo CD $ARGOCD_CHART_VERSION"
  helm --kube-context "$CONTEXT" upgrade --install argocd argo/argo-cd \
    --version "$ARGOCD_CHART_VERSION" \
    --namespace argocd --create-namespace \
    -f "$ARGO_DIR/install/values-local.yaml" \
    --wait --timeout 10m

  echo "==> projects, add-ons and the local root app"
  # Argo Workflows' CRD hook runs before Argo CD would create its namespace,
  # and its executor RBAC goes into ci: create both up front.
  for ns in argo ci; do
    kc create namespace "$ns" --dry-run=client -o yaml | kc apply -f - >/dev/null
  done
  # Applied directly once so add-ons (Helm-repo sources) start immediately;
  # from then on store-root-local keeps the same objects in sync from git.
  kc apply -n argocd -f "$ARGO_DIR/projects.yaml"
  kc apply -n argocd -f "$ARGO_DIR/addons/local/"
  kc apply -n argocd -f "$ARGO_DIR/root-local.yaml"

  echo "==> local CI (registry, pipeline)"
  # Applied from the working tree so builds work before the first push;
  # Application store-ci adopts the same objects once deploy/ci is on master.
  until kc get crd workflowtemplates.argoproj.io cronworkflows.argoproj.io \
        imageupdaters.argocd-image-updater.argoproj.io >/dev/null 2>&1; do
    echo "   waiting for Argo Workflows / Image Updater CRDs..."
    sleep 10
  done
  # Server-side: no client-side OpenAPI download (times out on a busy laptop API server).
  kc apply --server-side --field-manager=bootstrap-local -k "$ROOT/deploy/ci"
  kc -n ci rollout status deployment/registry --timeout=5m
  docker build -q -t localhost:5000/ci-tools:1 "$ROOT/deploy/ci/tools"
  docker push -q localhost:5000/ci-tools:1

  status
  cat <<EOF

Argo CD UI:   http://argocd.store.local:8081   (or: kubectl -n argocd port-forward svc/argocd-server 8080:80)
Pipelines:    http://workflows.store.local:8081  (Argo Workflows, namespace ci)
User:         admin
Password:     $(kc -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)

Add to /etc/hosts for the dev ingresses (ingress-nginx listens on :8081 / :8443):
  127.0.0.1 argocd.store.local workflows.store.local live.dev.store.example.com rooms.dev.store.example.com videos.dev.store.example.com live-api.dev.store.example.com vending.dev.store.example.com

store-dev and the pipeline follow GitHub master: they pick up the chart/,
deploy/ and service changes once those are pushed.
EOF
}

# App charts now run in store-dev; their old copies in `default` (rendered by
# the umbrella chart) only compete for CPU. Removes workloads, Services, HPAs,
# Ingresses and Jobs of every chart disabled in chart/values-infra-only.yaml.
# Secrets, ConfigMaps and PVCs are left in place.
remove_default_apps() {
  require_context
  local tmp
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' RETURN
  cp -R "$ROOT/chart" "$tmp/chart"
  (cd "$tmp/chart" && make -s k8s_subchart_deps >/dev/null && helm dependency update . >/dev/null)
  helm template "$UMBRELLA_RELEASE" "$tmp/chart" -n default > "$tmp/full.yaml"
  helm template "$UMBRELLA_RELEASE" "$tmp/chart" -n default -f "$tmp/chart/values-infra-only.yaml" > "$tmp/infra.yaml"
  python3 - "$tmp/full.yaml" "$tmp/infra.yaml" > "$tmp/targets.txt" <<'PY'
import sys, yaml
kinds = {"Deployment", "StatefulSet", "Service", "HorizontalPodAutoscaler", "Ingress", "Job", "CronJob"}
def objs(p):
    return {(d["kind"], d["metadata"]["name"]) for d in yaml.safe_load_all(open(p)) if d and d["kind"] in kinds}
for kind, name in sorted(objs(sys.argv[1]) - objs(sys.argv[2])):
    print(f"{kind.lower()}/{name}")
PY
  echo "Removing $(wc -l < "$tmp/targets.txt" | tr -d ' ') objects from namespace default:"
  sed 's/^/  /' "$tmp/targets.txt"
  xargs kubectl --context "$CONTEXT" -n default delete --ignore-not-found --wait=false < "$tmp/targets.txt"
}

# Start a pipeline run now instead of waiting for the next poll. Without an
# argument it builds what changed since the last run; "all" or a
# comma-separated list of services forces those builds.
ci_run() {
  require_context
  local force="${1:-}"
  kc -n ci create -o name -f - <<WF
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: store-ci-manual-
  namespace: ci
spec:
  workflowTemplateRef:
    name: store-ci
  arguments:
    parameters:
      - name: force
        value: "${force}"
WF
}

status() {
  require_context
  echo
  kc -n argocd get applications -o custom-columns=APP:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status 2>/dev/null || true
  echo
  kc -n ci get workflows --sort-by=.metadata.creationTimestamp 2>/dev/null | tail -5 || true
}

case "${1:-}" in
  install) install ;;
  remove-default-apps) remove_default_apps ;;
  ci-run) ci_run "${2:-}" ;;
  status) status ;;
  *) echo "usage: $0 {install|remove-default-apps|ci-run [all|svc,...]|status}" >&2; exit 2 ;;
esac
