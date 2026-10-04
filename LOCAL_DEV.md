
# SOUL Local Development

## 1. One-command goal

Command:
~~~bash
./scripts/soul-local-up.sh
~~~

It must:
1. verify Docker/Compose;
2. fetch or reuse exact SOUL repository revisions;
3. validate contract/config schemas;
4. generate untracked local secrets;
5. start NATS/Postgres/OTel/Prometheus/Grafana/Tempo;
6. start N01–N07 + SARA;
7. wait for readiness;
8. run discovery;
9. run authenticated seed capability transaction;
10. emit evidence summary.

Any missing executable surface causes an explicit BLOCKED result. Mocks are not used to create PASS evidence.

## 2. Local host ports

N01 8080  
N02 3020  
N03 3030  
N04 3040  
N05 3050  
N06 3060  
N07 8087  
SARA 8090  
NATS 4222  
PostgreSQL 5432  
OTLP HTTP 4318  
Prometheus 9090  
Grafana 3001  
Tempo 3200

## 3. Local secrets

File:
.env.soul.local

Ignored by git.

Minimum:
SOUL_MESH_SECRET
SOUL_MESH_TOKEN
SARA_API_TOKEN
N07_APP_TOKEN
provider keys needed by the selected real capability test

## 4. Seed scenario

conversation.start → N05 → document/tool operation in N04 → optional audio operation N03 → cognition N06 → N07 aggregation/validation → SARA trace.

Provider-dependent branches report DEGRADED when dependencies are absent.

## 5. Teardown

~~~bash
./scripts/soul-local-down.sh
~~~

Normal teardown does not wipe database volumes or evidence. Reset is an explicit separate command.

