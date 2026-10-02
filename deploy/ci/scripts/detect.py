#!/usr/bin/env python3
"""First step of the in-cluster CI pipeline (deploy/ci/workflow-template.yaml).

Keeps a git mirror of the repository on the workspace volume, works out which
services changed since the last processed commit, exports each one's build
context, and writes the build list for the next steps:

  /tmp/builds.json   [{name, image, tag, src, lang, goVersion, testEnv, db, buildArgs}]
  /tmp/sha           commit being built

Pipeline state (the last processed commit) lives in ConfigMap ci-state, so a
failed build is not retried every minute; pushing a fix rebuilds the service.

Environment:
  REPO_URL, BRANCH        what to follow
  WORKSPACE               mounted volume (mirror, exported sources, caches)
  RUN_ID                  unique per workflow run (export directory)
  FORCE                   "" | "all" | "svc-a,svc-b"  (manual runs)
"""
import json
import os
import subprocess
import sys

import yaml

REPO_URL = os.environ["REPO_URL"]
BRANCH = os.environ.get("BRANCH", "master")
WORKSPACE = os.environ.get("WORKSPACE", "/workspace")
RUN_ID = os.environ["RUN_ID"]
FORCE = os.environ.get("FORCE", "").strip()
MIRROR = f"{WORKSPACE}/repo.git"
STATE_CM = "ci-state"
REGISTRY = "localhost:5000"


def sh(*args, **kw):
    return subprocess.run(args, check=True, text=True, capture_output=True, **kw).stdout.strip()


def log(msg):
    print(msg, flush=True)


def sync_mirror():
    if not os.path.isdir(MIRROR):
        log(f"cloning {REPO_URL} (first run, the repository is large)...")
        subprocess.run(["git", "clone", "--mirror", REPO_URL, MIRROR], check=True)
    else:
        subprocess.run(["git", "-C", MIRROR, "remote", "update", "--prune"], check=True,
                       stdout=subprocess.DEVNULL)
    return sh("git", "-C", MIRROR, "rev-parse", f"refs/heads/{BRANCH}")


def get_state():
    out = subprocess.run(["kubectl", "get", "configmap", STATE_CM, "-o", "jsonpath={.data.lastSha}"],
                         text=True, capture_output=True)
    return out.stdout.strip() if out.returncode == 0 else ""


def set_state(sha):
    patch = json.dumps({"data": {"lastSha": sha}})
    if subprocess.run(["kubectl", "patch", "configmap", STATE_CM, "--type", "merge", "-p", patch],
                      capture_output=True).returncode != 0:
        subprocess.run(["kubectl", "create", "configmap", STATE_CM, f"--from-literal=lastSha={sha}"],
                       check=True, capture_output=True)


def catalog_at(sha):
    try:
        return yaml.safe_load(sh("git", "-C", MIRROR, "show", f"{sha}:deploy/ci/services.yaml"))["services"]
    except subprocess.CalledProcessError:
        return []


def main():
    sha = sync_mirror()
    last = get_state()
    services = catalog_at(sha)
    log(f"head {sha[:7]}, last processed {last[:7] or '-'}, {len(services)} services in catalog")

    if FORCE:
        wanted = {s["name"] for s in services} if FORCE == "all" else set(FORCE.split(","))
        selected = [s for s in services if s["name"] in wanted]
        unknown = wanted - {s["name"] for s in services}
        if unknown:
            log(f"unknown services ignored: {', '.join(sorted(unknown))}")
    elif not last:
        # First run: adopt the current commit as the baseline. store-dev keeps
        # the images it already runs until a service's code changes (or a
        # manual run with FORCE builds it).
        log("no previous state, recording baseline without building")
        selected = []
    elif last == sha:
        selected = []
    else:
        changed = sh("git", "-C", MIRROR, "diff", "--name-only", last, sha).splitlines()
        selected = [s for s in services if any(f.startswith(s["context"].rstrip("/") + "/") for f in changed)]
        log(f"{len(changed)} files changed")

    if not FORCE:
        set_state(sha)

    tag = sh("git", "-C", MIRROR, "show", "-s", "--format=%cd", "--date=format-local:%Y%m%dT%H%M%S", sha,
             env={**os.environ, "TZ": "UTC"}) + "-" + sha[:7]
    builds = []
    for s in selected:
        src = f"{WORKSPACE}/src/{RUN_ID}/{s['name']}"
        os.makedirs(src, exist_ok=True)
        archive = subprocess.Popen(["git", "-C", MIRROR, "archive", "--format=tar", f"{sha}:{s['context']}"],
                                   stdout=subprocess.PIPE)
        subprocess.run(["tar", "-x", "-C", src], stdin=archive.stdout, check=True)
        if archive.wait() != 0:
            raise SystemExit(f"git archive failed for {s['name']}")
        builds.append({
            "name": s["name"],
            "image": f"{REGISTRY}/{s['image']}",
            "tag": tag,
            "src": src,
            "lang": s.get("lang", "other"),
            "goVersion": str(s.get("goVersion", "1.25")),
            "testEnv": " ".join(f"{k}={v}" for k, v in (s.get("testEnv") or {}).items()),
            "db": "true" if s.get("testServices") else "false",
            "buildArgs": " ".join(f"--build-arg={k}={v}" for k, v in (s.get("buildArgs") or {}).items()),
        })
        log(f"  build {s['name']} -> {REGISTRY}/{s['image']}:{tag}")

    with open("/tmp/builds.json", "w") as f:
        json.dump(builds, f)
    with open("/tmp/sha", "w") as f:
        f.write(sha)
    log(f"{len(builds)} service(s) to build")


if __name__ == "__main__":
    sys.exit(main())
