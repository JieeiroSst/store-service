# Security Policy

This document describes how security is handled for the **store-service** monorepo: which code is covered, how to report a vulnerability, what happens after you report it, and the rules contributors must follow when handling secrets.

## Supported Versions

store-service follows a rolling-release model. There are no long-term release branches; only the latest code on `master` and the images built from it receive security fixes.

| Version / Artifact                                   | Supported          |
| ---------------------------------------------------- | ------------------ |
| `master` branch (latest commit)                      | :white_check_mark: |
| Container images built from the latest `master` CI   | :white_check_mark: |
| Umbrella Helm chart `app-cluster` `0.1.x`            | :white_check_mark: |
| Service sub-charts `0.2.x`                           | :white_check_mark: |
| Service sub-charts `< 0.2.0`                         | :x:                |
| Forks, feature branches, older image tags            | :x:                |

If you run an older image or chart, upgrade to the latest `master` build before reporting. The issue may already be fixed.

## Scope

### In scope

- **Application services**: every service directory at the repository root (for example `user_service`, `payment_service`, `wallet-service`, `kms-service`, `sso-service`, `authorize_service`, `ekyc-service`, `banking-service`, `sard-service`).
- **Gateways**: `gateway-api` (GraphQL gateway) and `gateway-internal-api`, including authentication, authorization and proxy/routing logic.
- **Deployment and infrastructure as code**:
  - `chart/`: the `app-cluster` umbrella Helm chart, service sub-charts, and the shared `common` and `platform` charts (including the `ExternalSecret` / `SecretStore` templates).
  - `deploy/`: Argo CD applications, ApplicationSets, projects, add-ons and the in-cluster CI workflows.
  - `.github/workflows/`: CI/CD pipelines, including image build, chart CI and promotion.
- **Default configuration** shipped in the repository that would make a deployment insecure (exposed admin ports, default credentials, overly permissive RBAC or network policies).

### Out of scope

- Vulnerabilities in third-party components (Kafka, Keycloak, PostgreSQL, Redis, MinIO, Elasticsearch, Ollama, and so on). Report these upstream. Tell us if our *configuration* of a component makes it exploitable.
- Findings that only affect a local development environment and cannot be reached in a deployed cluster.
- Missing best-practice headers or banners with no demonstrated impact.
- Denial-of-service from volumetric or brute-force traffic.
- Social engineering, physical attacks, and attacks on contributors' personal accounts.
- Reports produced only by automated scanners and not verified as exploitable.

## Reporting a Vulnerability

**Do not open a public GitHub issue, pull request or discussion for security vulnerabilities.**

Report privately through either channel:

1. **GitHub private vulnerability reporting (preferred)**: go to the repository's **Security** tab and choose **Report a vulnerability**.
   <https://github.com/JieeiroSst/store-service/security/advisories/new>
2. **Email**: **luumanhquan.91@gmail.com**, with the subject line `[SECURITY] store-service: <short summary>`.

Include as much of the following as you can:

- Affected service(s), chart(s) or workflow(s), with file paths and the commit SHA or image tag
- Vulnerability type (for example auth bypass, IDOR, SQL injection, SSRF, secret exposure, RCE)
- Step-by-step reproduction instructions or a proof of concept
- Impact: what an attacker could achieve
- Any suggested fix or mitigation
- Whether and how you want to be credited

Do **not** include real customer data, live credentials or secrets in your report. If you found a leaked secret, describe where it is (file and commit) without copying its value.

### Response timeline

| Stage                                  | Target time                          |
| -------------------------------------- | ------------------------------------ |
| Acknowledgement of your report         | within **48 hours**                  |
| Initial triage and severity assessment | within **5 business days**           |
| Status updates                         | at least every **7 days** until resolved |
| Fix for Critical severity              | within **7 days** of confirmation    |
| Fix for High severity                  | within **30 days** of confirmation   |
| Fix for Medium / Low severity          | within **90 days** of confirmation   |

Severity is assessed with [CVSS v3.1](https://www.first.org/cvss/calculator/3.1).

### Handling process

1. **Receive**: we acknowledge your report and assign a tracking ID, usually a private GitHub Security Advisory.
2. **Triage**: we reproduce the issue, confirm scope and assign a severity.
3. **Decide**:
   - **Accepted**: we tell you the severity, the planned fix timeline and the advisory ID.
   - **Declined**: we explain why (for example out of scope, not reproducible, or intended behavior). You may reply with more information and ask us to reconsider.
4. **Fix**: we develop the patch in a private fork or advisory branch. We may ask you to verify it.
5. **Release**: we merge to `master`, CI rebuilds the affected images, and Argo CD rolls them out. Rotated secrets are replaced through the secret store.
6. **Disclose**: we publish the advisory as described below.

## Disclosure Policy

We follow **coordinated disclosure**:

- Please give us a reasonable time to fix the issue before disclosing it publicly: **90 days** from the report, or less if a fix is released earlier.
- After a fix is released, we publish a GitHub Security Advisory with a description, affected components, severity, fixed commit and credit to the reporter (unless you prefer to stay anonymous). We request a CVE when appropriate.
- If a vulnerability is being actively exploited, we may disclose earlier, together with mitigation guidance.
- If we cannot meet the 90-day deadline, we will contact you before it passes to agree on an extension.

## Safe Harbor

We will not pursue legal action against good-faith security research that:

- Follows this policy and reports privately,
- Avoids privacy violations, data destruction and service disruption,
- Only accesses the minimum data needed to demonstrate the issue, and deletes it afterwards,
- Does not exploit the issue beyond what is needed to confirm it.

Only test against your own deployment of this project, never against production systems you do not own.

## Secret Management

These rules apply to every contributor and every service in this repository.

### Never commit secrets

- Never commit `.env`, `.env.local`, `.env.*.local`, private keys, kubeconfigs, cloud credentials, API keys (for example payment gateways, SerpApi, Solana or Shopify keys) or database passwords. These files are already listed in `.gitignore`.
- Commit a `.env.example` with placeholder values to document the variables a service needs.
- Do not hard-code secrets in source code, `config/` files, Helm `values.yaml`, Argo CD manifests, Dockerfiles or GitHub workflow files.
- Do not print secrets in logs, error messages, CI output or API responses.

### How secrets are provided

- **Kubernetes**: secrets come from the External Secrets Operator. Charts reference them through the shared `common` templates (`_externalsecret.tpl`) and the `SecretStore` defined in `chart/platform`. Plain `Secret` manifests with real values must never be committed.
- **Encryption keys**: application-level encryption and signing keys are managed by `kms-service`. Services must request keys from it instead of storing raw key material in their own configuration or database.
- **CI/CD**: use GitHub Actions encrypted secrets or environment secrets. Never echo them, and never pass them to workflows triggered from forks.

### If a secret is leaked

1. **Rotate or revoke the secret immediately.** Deleting the commit is not enough, because the value stays in git history and in any clones.
2. Update the new value in the secret store or in the GitHub secrets.
3. Remove the value from git history (for example with `git filter-repo`) if needed, and tell the maintainers.
4. Report the incident through the private channels above so that its impact can be assessed.

## Security Best Practices for Contributors

- Keep dependencies up to date (`go.mod`, `package.json`, `requirements.txt`, `Cargo.toml`, `build.sbt`) and fix known-vulnerable versions promptly.
- Validate and sanitize all input at service boundaries. Use parameterized queries only.
- Enforce authentication and authorization at `gateway-api` **and** inside each service. Do not rely on the gateway alone.
- Use TLS for external traffic, and do not expose internal services (databases, Kafka, Consul, admin UIs) outside the cluster.
- Run containers as non-root with minimal base images, and keep RBAC and network policies least-privilege.

---

Thank you for helping keep store-service and its users safe.
