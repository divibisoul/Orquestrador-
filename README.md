# ORQUESTRADOR-NEXUS

Plataforma cognitiva distribuída substrate-agnostic. A fundação implementa três pilares: **Orquestrador**, **Neo-Córtex Pré-Frontal** e **Super AGI**, ligados por uma malha de capacidades.

## Estado da fundação

Esta branch (`foundation/orchestrator-nexus-v1`) contém uma implementação executável sem dependência obrigatória de fornecedor: workflow/DAG validation, execução paralela/distribuída, resiliência, registro de nós, roteamento por capacidade, estado versionado, planejamento executivo, memória e uma camada de modelo plugável com fallback determinístico.

As metas de 15 ms P95, 1.200 RPS, 40% de economia energética, 99,99% e 1.000 nós são **SLOs de benchmark**, não capacidades declaradas sem medição.

## Fluxo

```text
Objetivo -> Neo-Córtex -> Plano/DAG -> Orquestrador -> Mesh -> Nó/Modelo
                         ^                         |
                         |---- avaliação/feedback-|
```

## Executável

`go run ./cmd/nexus` inicia o plano de controle HTTP em `:8080`.

- `GET /health` — saúde do processo e contagem de nós
- `GET /v1/nodes` — snapshot da federação
- `POST /v1/plan` — gera e decide um plano a partir de `{ "goal": "..." }`

## Arquitetura

Consulte [ARCHITECTURE.md](ARCHITECTURE.md), [API_REFERENCE.md](docs/API_REFERENCE.md), [DEPLOYMENT.md](docs/DEPLOYMENT.md) e [SECURITY.md](docs/SECURITY.md).

## Princípios

1. Sem lock-in de modelo, hardware ou transporte.
2. Capacidades são anunciadas por nós e usadas por matching dinâmico.
3. Falhas são estados explícitos; não são mascaradas como sucesso.
4. Infraestrutura externa (Raft, GPU/NPU, mTLS, OpenTelemetry) entra por boundaries/adapters, preservando o núcleo compilável e testável.
5. Toda alegação de performance deve ser comprovada por benchmark reproduzível.
