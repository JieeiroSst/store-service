# Triển khai production (GitOps)

Thiết kế triển khai store-service theo mô hình production: mỗi môi trường là một
namespace riêng, Argo CD đồng bộ trạng thái từ git, CI không bao giờ `kubectl apply`.

```
 push master ──► GitHub Actions ──► build + test + Trivy scan ──► ghcr.io/<image>:<sha>
                                         │
                                         └─► commit tag vào chart/<svc>/values-dev.yaml
                                                     │
 Argo CD (deploy/argocd) ◄───────── theo dõi git ────┘
   ├── store-dev      auto-sync
   ├── store-staging  auto-sync   ◄── PR "Promote image" dev → staging
   └── store-prod     sync tay    ◄── PR "Promote image" staging → prod
```

## Thành phần

| Đường dẫn | Vai trò |
|---|---|
| `chart/common` | Library chart: Deployment, Service, Ingress (TLS), HPA, PDB, NetworkPolicy, ServiceAccount, ServiceMonitor, ExternalSecret theo chuẩn production. |
| `chart/platform` | Rào chắn cho từng namespace: Namespace + Pod Security Admission, ResourceQuota, LimitRange, PriorityClass, NetworkPolicy nền (default-deny, DNS, Consul, ingress controller, Prometheus), SecretStore. |
| `chart/<svc>/values-{dev,staging,prod}.yaml` | Overlay theo môi trường. Có file này = service tham gia GitOps cho môi trường đó. |
| `deploy/argocd` | AppProject (`store-nonprod`, `store-prod`, `store-addons`), Application/ApplicationSet theo từng môi trường, add-on, app-of-apps `root.yaml` (production) / `root-local.yaml` (laptop), values cài Argo CD. |
| `deploy/bootstrap-local.sh` | Dựng toàn bộ GitOps trên docker-desktop. |
| `.github/workflows/promote.yml` | Mở PR chuyển image tag dev → staging → prod. |
| `.github/workflows/chart-ci.yml` | Lint + render mọi chart × môi trường, validate bằng kubeconform. |

## Chuẩn production mà `chart/common` áp dụng

- **Rollout an toàn**: `RollingUpdate` với `maxUnavailable: 0`, startup/readiness/liveness probe, `preStop` sleep để rút endpoint trước SIGTERM, `terminationGracePeriodSeconds` cấu hình được.
- **Sẵn sàng cao**: HPA (CPU + memory, scale-down chậm 5 phút), PDB, topology spread theo node và zone, PriorityClass.
- **Bảo mật**: chạy non-root, `readOnlyRootFilesystem`, drop mọi capability, seccomp `RuntimeDefault`, ServiceAccount riêng không mount token, namespace enforce PSA `restricted` (dev: `baseline`).
- **Mạng zero-trust**: default-deny ở namespace; service khai báo rõ ai được gọi nó (`allowFromApps`) và nó gọi ai (`allowToApps`, `egress`). Các quyền chung gắn bằng label `network.store.io/*`.
- **Secrets**: không để mật khẩu trong git. `ExternalSecret` lấy từ `SecretStore` `store-secrets` — non-prod mirror Secret do chart postgres tạo trong `default`, prod đọc từ Vault.
- **Image bất biến**: prod dùng tag commit SHA (hoặc `image.digest`), không dùng `latest`; image có SBOM + provenance và phải qua Trivy (CRITICAL có bản vá ⇒ chặn).

## Luồng CI/CD tự động

```
git push master
  └─ CI trong cluster (deploy/ci, Argo Workflows, namespace ci) — mỗi phút poll GitHub
       detect: fetch git mirror, diff với commit đã xử lý → các service có thay đổi (deploy/ci/services.yaml)
       test:   go vet + go test (Postgres/MySQL sidecar khi service cần)            ┐ 2 service song song
       image:  docker build trên Docker daemon của node → push localhost:5000/<image>:<giờ commit>-<sha7> ┘
  └─ Argo CD Image Updater (deploy/ci/image-updater.yaml) thấy tag mới (~30s) → set image cho <svc>-dev
       └─ Argo CD sync store-dev → rolling update
staging / prod: Actions → "<Service> CI/CD" (chạy tay, build ghcr.io) → "Promote image" (service, tag) → PR → merge
```

- GitHub Actions không còn tự chạy khi push (chỉ `chart-ci` trên Pull Request). Workflow của từng service chỉ dùng để phát hành image GHCR cho staging/prod.
- Build tay: `deploy/bootstrap-local.sh ci-run vending-machine-service` (hoặc `all`); xem pipeline ở `http://workflows.store.local:8081`.
- Thêm service vào CI: thêm một mục vào `deploy/ci/services.yaml`, rồi chạy `python3 deploy/ci/scripts/gen-image-updater.py`.
- Image tag của dev là trạng thái trong cluster (tham số Helm do Image Updater đặt trên Application), không commit vào git; staging/prod đi qua git.
- Repo được clone dạng mirror trên PVC `ci-workspace`; lần đầu mất ~10 phút.

## Môi trường local (docker-desktop)

Laptop chạy **store-dev** do Argo CD quản lý. Hạ tầng dùng chung (Consul, Postgres, MySQL, Kafka, ...) vẫn chạy bằng umbrella chart trong `default`, và platform chart tạo alias ExternalName tới chúng.

```sh
deploy/bootstrap-local.sh install              # Argo CD + ingress-nginx + External Secrets + metrics-server + store-root-local
deploy/bootstrap-local.sh remove-default-apps  # xóa bản sao app cũ trong `default` (giữ Secret/ConfigMap/PVC)
deploy/bootstrap-local.sh status

# Hạ tầng trong `default`, không kèm app:
helm template chart-1785009827 ./chart -n default -f chart/values-infra-only.yaml | kubectl apply -n default -f -
```

- `deploy/argocd/root-local.yaml` chỉ lấy `projects.yaml`, `applicationsets/*-dev.yaml` và `addons/local/`. Cluster production dùng `root.yaml`, gồm staging, prod và `addons/production/` (có thêm cert-manager và ClusterIssuer).
- ingress-nginx local nghe ở **:8081 / :8443**, vì cổng 80 đã bị `nginx-loadbalancer` của umbrella chiếm. Thêm các host vào `/etc/hosts` (script in sẵn danh sách), rồi mở `http://argocd.store.local:8081` hoặc `http://live.dev.store.example.com:8081`.
- Overlay dev đặt requests 10m CPU / 64Mi cho mỗi service (limits giữ nguyên), để ~70 service vừa một node 8 CPU.
- Ở local không có cert-manager: ingress dùng chứng chỉ mặc định của nginx, và annotation `cert-manager.io/cluster-issuer` bị bỏ qua.
- Repo nặng ~5 GB nên lần clone đầu của Argo CD mất khoảng 10 phút; các lần sau chỉ fetch phần thay đổi. Về lâu dài nên tách chart và `deploy/` sang một repo GitOps riêng, nhỏ.

## Cài đặt cluster production

Cài Argo CD (có thể dùng `deploy/argocd/install/values-local.yaml` làm điểm xuất phát, bật HA), rồi:

```sh
kubectl apply -n argocd -f deploy/argocd/projects.yaml
kubectl apply -n argocd -f deploy/argocd/root.yaml
```

Add-on (ingress-nginx, cert-manager, External Secrets, metrics-server) được cài từ `deploy/argocd/addons/production`. Trước khi bật prod, sửa các giá trị giả định:

- `deploy/argocd/addons/production/cluster-issuers.yaml` → email ACME.
- `chart/platform/values.yaml` → `secretStore.providerSpec` (địa chỉ Vault/role), `infraAliases` (endpoint của DB/broker được quản lý).
- `chart/<svc>/values-prod.yaml` → host ingress, CIDR database được quản lý, đường dẫn secret.
- `deploy/argocd/projects.yaml` → group SSO `store-release-managers`, lịch change freeze.

Cần bật **Settings → Actions → Allow GitHub Actions to create pull requests** để workflow promote mở PR. Nếu `master` có branch protection, cho phép `github-actions[bot]` push (bước bump tag dev).

## Quy trình phát hành

1. Push vào `master` → CI trong cluster test + build → Image Updater → Argo CD deploy `store-dev`.
2. Actions → **<Service> CI/CD** (Run workflow) → image `ghcr.io/jieeirosst/<image>:<sha7>` (đã qua Trivy).
3. Actions → **Promote image** (`service`, `tag`, `staging`) → review + merge PR → Argo CD deploy `store-staging`.
4. **Promote image** với cùng tag tới `prod` (chỉ nhận đúng tag staging đang chạy) → merge PR → release manager bấm **Sync** app `<svc>-prod` (ngoài khung freeze thứ Sáu 12:00 UTC – thứ Hai).
5. Rollback: revert PR promote (hoặc Argo CD → History → Rollback, sau đó revert git để không bị self-heal ghi đè).

## Trạng thái các chart

Tất cả 66 chart ứng dụng chạy trên `chart/common` và có overlay dev/staging/prod, kể cả các chart có cấu trúc đặc biệt:

- **livestream-service**: node là StatefulSet có sidecar SRS (`statefulSet`, `sidecarsTemplate`), edge là Deployment. Địa chỉ RTMP/HTTP của từng pod dùng `{{ .Release.Namespace }}` thay vì hardcode `default`.
- **airflow-server**: webserver (`common.app`), scheduler (`common.workload`), postgres metadata. Mật khẩu, chuỗi kết nối DB và fernet key nằm trong Secret `airflow-server-secrets`, không còn nằm trong env hay args. Ở các môi trường Argo CD, postgres có volume (local vẫn không có, giống trước).
- **geoservice-postgres**: PostGIS riêng cho geoservice. Mật khẩu chuyển từ ConfigMap sang Secret, strategy `Recreate`, PVC được giữ lại khi xóa release.

Chỉ các chart hạ tầng dùng chung (mysql, postgres, kafka, consul, ...) vẫn chạy bằng umbrella chart.

Mặc định của overlay (sinh theo cùng một mẫu, chỉnh tay khi cần):

| | dev | staging | prod |
|---|---|---|---|
| Service type | ClusterIP | ClusterIP | ClusterIP |
| HPA min–max | 1–3 | 2–4 | 3–10 |
| PDB, topology spread | – | có | có |
| Pod Security | non-root uid 10001, drop ALL | như dev | như dev + PriorityClass |
| Secrets | Secret do chart render từ values | như dev | ExternalSecret (`store/prod/<svc>`, `store/prod/<svc>/env-file`) |
| Ingress | livestream-web, room-web, video-web, livestream edge API (và vending) | như dev | như dev |

Các service còn lại chỉ nhận traffic nội bộ, qua NetworkPolicy `allow-intra-namespace`. Muốn mở ra ngoài thì đặt `ingress.enabled` và `ingress.host` trong overlay.

## Contract values của `chart/common`

Mỗi khóa service trong `values.yaml` (ví dụ `paymentService:`) được truyền vào `common.app` (service HTTP) hoặc `common.workload` (worker không có Service). Các khóa chính:

| Khóa | Ý nghĩa |
|---|---|
| `selectorApp` | Nhãn `app` dùng cho selector. Phần lớn chart dùng `<name>-deployment`; **không được đổi** vì selector của Deployment là bất biến. |
| `image.port`, `extraPorts` | Port chính (tên `http`) và các port phụ, ví dụ `grpc`. |
| `service`, `service.extraPorts`, `extraServices` | Service chính, port phụ trên cùng Service, và các Service riêng (ví dụ `*-grpc-svc`). |
| `ingress`, `extraIngresses` | Ingress chính và các Ingress phụ (ví dụ host gRPC). |
| `containerEnv` | Danh sách env. `value` được render qua `tpl` nên có thể trỏ tới khóa khác (`"{{ .Values.x.env.mysqlHost }}"`). Mục có `when` chỉ được render khi biểu thức đó khác rỗng. |
| `envFile` | File `.env` mount vào `/app/.env` (Secret `<name>-env-file`). Ở prod lấy từ secret manager qua `envFile.externalSecret`. |
| `volumes`, `volumeMounts` | Volume thêm, cũng hỗ trợ `tpl` và `when`. |
| `initContainersTemplate` | Tên một `define` trong `templates/_init-containers.tpl` của chart (giữ nguyên logic template của init container). |
| `probes` | startup/readiness/liveness. Service chưa có endpoint health thì dùng TCP trên port `http`. |
| `consulSeed` | Seed key Consul (`key`) từ `files/consul.json` của chart nếu key chưa có. Với Argo CD là Sync hook chạy trước Deployment; với helm là post-install. |
| `externalSecret` | Đồng bộ Secret của chart (`targetName`) từ `SecretStore` `store-secrets`. Khi bật, `templates/secret.yaml` của chart sẽ tắt. |
| `deploymentName` | Tên workload, mặc định `<name>-deployment`. Chart có object cũ mang tên khác thì giữ tên đó, để upgrade thay pod tại chỗ. |
| `statefulSet.serviceName` | Render StatefulSet thay vì Deployment (DNS cố định cho từng pod qua headless Service). |
| `sidecars`, `sidecarsTemplate` | Container chạy kèm trong pod; tự nhận `securityContext` của service. |
| `networkLabels.isolated` | Loại pod khỏi quy tắc "mọi pod trong namespace gọi nhau"; chỉ caller trong `networkPolicy.allowFromApps` mới kết nối được (dùng cho database). |

Khi viết chart mới, xem `chart/vending-machine-service` (service HTTP), `chart/video-service` (API + worker dùng chung image), `chart/livestream-service` (StatefulSet + sidecar) và `chart/geoservice-postgres` (database).

## Hạn chế đã biết

- **Tài nguyên laptop**: Docker Desktop 8 CPU / 16 GB đang quá tải ngay cả khi không có app nào chạy (load ~100, kube-apiserver ~130% CPU, kube-scheduler đã restart hơn 450 lần) do hạ tầng trong `default` (elasticsearch, kibana, prometheus, loki/promtail, kafka-connect, keycloak, rabbitmq...). `store-dev` đang để 0 replica. Muốn chạy toàn bộ cần tăng tài nguyên cho Docker Desktop hoặc tắt bớt hạ tầng không dùng; bật service theo nhóm: `kubectl -n store-dev scale deploy <tên> --replicas=1` (HPA sẽ tự quản lý tiếp).
- **Consul** trước đây chạy `agent -dev` (lưu KV trong RAM), nên mỗi lần restart mất toàn bộ cấu hình, kể cả các bản backup sau ngày 26/9. Đã chuyển sang server mode, lưu trên PVC. KV được seed lại từ `consul.json` tham chiếu trong repo; các giá trị đã chỉnh tay trước đây không khôi phục được.
- voucher-service không có `consul.json` tham chiếu, nên chưa được seed.

- **Non-root**: phần lớn image không khai báo `USER`. Overlay chạy chúng với uid 10001 và cho phép ghi root filesystem (`readOnlyRootFilesystem: false`). Service nào ghi vào thư mục thuộc root (ví dụ `/app/...`) sẽ lỗi ở dev; khi đó cần thêm `USER` vào Dockerfile, hoặc mount `emptyDir` vào đường dẫn đó.
- **Probe TCP** được thêm cho các service trước đây không có probe. Nếu `image.port` không khớp port thật mà app đọc từ Consul, pod sẽ bị restart (trước đây lỗi này bị che đi).
- **room-web, video-web** chuyển sang `nginx-unprivileged` và nghe port 8080. Image `:latest` cũ vẫn nghe port 80, nên cần build lại image (CI tự build khi merge) trước khi deploy chart mới.
- 19 chart mount `/app/.env` từ Secret dùng chung `service-env-files`, nhưng Secret đó không có key cho các service này (tình trạng có từ trước). Volume đã được đánh dấu `optional` để pod vẫn chạy được ở namespace `store-*`. Nếu service thực sự cần file `.env`, hãy thêm `envFile` cho nó.
- livestream-web, threads-service, ekyc-service và doordash-service trước đây gọi `http://user-service-svc`, một Service không tồn tại. Giờ đã trỏ tới `http://user-api-svc` (REST API của chart `user`), cả trong chart, file `consul.json` tham chiếu, lẫn giá trị mặc định trong code. Riêng các key Consul KV **đang chạy** không được seed job ghi đè, nên cần sửa tay `userService.baseURL` của `threads_service`, `ekyc_service`, `doordash_service`.
- **Cluster local** (deploy bằng `helm template | kubectl apply`): 3 Deployment của airflow-server chuyển env từ `value` sang `valueFrom` (Secret), nhưng `kubectl apply` không gỡ được trường `value` cũ trên object do helm tạo. Cần xóa chúng một lần trước khi apply: `kubectl -n default delete deployment airflow-server-webserver airflow-server-scheduler airflow-postgres-deployment`. Postgres local không có volume, nên dữ liệu metadata vốn đã mất mỗi lần restart.
- Consul KV đang dùng chung cho local/dev/staging (cùng instance trong `default`), nên cấu hình của các môi trường không tách biệt. Prod trỏ `consul-service` sang Consul riêng ở namespace `consul`.
- Nhãn `network.store.io/internet` (cho phép gọi HTTPS ra ngoài) được gán dựa trên việc quét source code. Service nào gọi API bên ngoài mà bị chặn thì bật `networkLabels.internet: true` trong overlay.
- Chart `postgres` sinh mật khẩu bằng `lookup`, hàm này luôn trả rỗng khi Argo CD render, nên nếu đưa postgres vào Argo CD thì mỗi lần sync sẽ sinh mật khẩu mới. Prod nên dùng database được quản lý (RDS/Cloud SQL) với mật khẩu nằm trong Vault.
- Image `minio/minio:latest` không còn tải được từ Docker Hub (minio trong `default` đang ImagePullBackOff), nên video/livestream không có object storage cho tới khi đổi image cho chart `minio`. Ngoài ra `mysql-river` và `temporal` đang CrashLoop (lỗi có từ trước).
- Hạ tầng stateful (Consul, Postgres, Kafka, ...) ở non-prod vẫn chạy bằng umbrella chart trong `default`. Platform chart tạo Service ExternalName cùng tên (`postgresdb`, `mysql-svc`, ...) trong từng namespace `store-*`, nên service vẫn gọi bằng tên ngắn. Ở prod, các alias này trỏ tới endpoint được quản lý (`chart/platform/values.yaml` → `infraAliases`).
- Bước tiếp theo nên làm: progressive delivery (Argo Rollouts canary + phân tích bằng Prometheus), ký image bằng cosign kèm policy admission (Kyverno), backup bằng Velero.
