# Load test: one room, a million users, several pods

The promise: however many users ask at once, on however many pods, exactly as many reservations succeed as there
are rooms. Two ways to check it.

## 1. In the test suite (no cluster needed)

`internal/integration` runs several independent replicas of the service (each with its own connection pool,
sold-out cache, bulkhead and background loops) against one real Postgres:

```
TEST_DATABASE_URL='postgres://postgres:pw@localhost:5432/postgres?sslmode=disable' go test ./internal/integration/ -v
MULTIPOD_USERS=1000000 MULTIPOD_WORKERS=5000 TEST_DATABASE_URL=... go test ./internal/integration/ -run OneRoom -v
```

Without `TEST_DATABASE_URL` the tests are skipped. CI runs them against a Postgres service container.

## 2. On a real cluster

Everything below lives in the namespace `rs-loadtest`; `kubectl delete namespace rs-loadtest` removes it all. It
needs about 8 CPUs of headroom for a million requests: use a cluster you can spare, not a shared one.

```
# from reservation-service/
docker build -f loadtest/Dockerfile -t rs-loadtest:test .          # the load generator and the user-service stand-in
docker build -t reservation-service:test .                         # the service
# (with a registry: push both and set image names/pullPolicy in loadtest/k8s/*.yaml and values.yaml)

kubectl apply -f loadtest/k8s/loadtest.yaml                       # namespace, Postgres, the user-service stand-in
helm template ../chart/reservation-service -f loadtest/k8s/values.yaml -n rs-loadtest | kubectl apply -n rs-loadtest -f -
kubectl -n rs-loadtest rollout status deploy/reservation-service-deployment

# one hotel, one room type, ONE room for one night (token "admin" is the stand-in's admin)
kubectl -n rs-loadtest run seed --rm -i --restart=Never --image=curlimages/curl --command -- sh -c '
  A="Authorization: Bearer admin"; B=http://reservation-service-svc/api/v1
  curl -s -XPOST $B/hotels -H "$A" -d "{\"name\":\"Grand\",\"city\":\"Da Nang\",\"address\":\"1 Beach\",\"phone_number\":\"+84 236 123 456\",\"currency\":\"VND\",\"stars\":5}"
  curl -s -XPOST $B/hotels/1/room-types -H "$A" -d "{\"name\":\"Suite\",\"capacity\":4}"
  curl -s -XPUT $B/hotels/1/room-types/1/inventory -H "$A" -d "{\"from\":\"2030-01-01\",\"to\":\"2030-01-03\",\"total_inventory\":1,\"rate\":20000000}"'

kubectl apply -f loadtest/k8s/job.yaml                            # the storm
kubectl -n rs-loadtest logs -f job/flash                          # exits non-zero unless exactly one reservation succeeded
kubectl -n rs-loadtest exec deploy/postgres -- psql -U testUser -d reservation_service -c \
  "select count(*) reservations, (select max(total_reserved) from room_type_inventory) reserved from reservation where status = 1"
```

What to look at besides the winner count:

- **Database connections** (`select count(*) from pg_stat_activity`): replicas x `postgresMaxConns` must stay under
  Postgres `max_connections` (100 by default). A pod that cannot get a connection answers 429 instead of 500, but it
  cannot serve until it does.
- **Probes**: a flooded pod answers health checks slowly. The chart's liveness probe is patient (5 s timeout, 6
  failures) so a busy pod is taken out of rotation for a moment instead of being restarted.
- **CPU limits**: a pod limited to 1 CPU serves about a fifth of what it does unlimited; that is the cost of the limit, not
  a bug.
