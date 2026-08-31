# ORQUESTRADOR-NEXUS

Plataforma cognitiva distribuída e provider-neutral. Esta recuperação consolida a fundação existente em uma base canônica, testável e preparada para evolução entre os seis planos.

## Arquitetura

```text
Objetivo
   |
   v
Neo-Córtex -> Neural Fabric -> Orquestrador -> Compute
   |              |               |
   +--------------+---------------+
                  |
              Mesh <-> State
                  |
             API / Protobuf
```

Pacotes canônicos de domínio:

- `core/orchestrator` — engine único de workflow, execução e resiliência.
- `core/prefrontal` — planejamento, decisão, contexto e feedback executivo.
- `core/neuralfabric` — encoding, previsão, seleção de rotas e feedback.
- `core/superagi` — geração/provider boundary, memória e verificação.
- `mesh` — registro, descoberta, heartbeat e stale-node.
- `state` — armazenamento versionado e CAS.
- `compute` — abstração de dispositivos e execução local/simulada.
- `security` — policy/auth/audit boundary.

As antigas implementações `internal/nexus`, `internal/mesh` e `internal/state` foram removidas da árvore de recuperação após verificação de consumidores. O histórico original continua preservado no Git e na branch `foundation/orchestrator-nexus-v1`.

## Maturidade

**Implementado:** workflow lifecycle, execução sequencial/paralela/distribuída, retry, circuit breaker, bulkhead, rate limiter, fractal primitives; planejamento/decisão executiva; roteamento Neural Fabric; registro/descoberta/heartbeat do Mesh; state versionado/CAS; políticas básicas de segurança; E2E dos seis planos.

**Parcial/boundary:** aprendizagem adaptativa/persistência da Neural Fabric; modelos/provider reais do Super AGI; GPU/NPU físico; Raft distribuído; mTLS/SPIFFE; observabilidade externa; runtime gRPC.

**Experimental:** adapter Gemini da branch `feature/gemini-provider`; permanece isolado do runtime Go canônico.

## Executar

Requisitos: Go conforme `go.mod` e Python 3 para os testes do adapter.

```bash
go mod tidy
go build ./...
go test ./...
go test ./... -race
go run ./cmd/nexus
```

Demonstração independente dos seis planos:

```bash
go run ./cmd/orchestrator
```

O plano de controle Nexus usa `:8080`. Consulte `docs/API_REFERENCE.md`.

## CI / self-hosted runner

A recuperação usa um GitHub Actions runner auto-hospedado para CI, segurança e benchmarks. O runner deve fornecer Linux/x64, compatibilidade com a versão configurada do Actions Runner, Go 1.23, Python 3 e Docker para o scanner Gosec.

Consulte `docs/SELF_HOSTED_RUNNER.md` para a preparação e verificação do ambiente.

## Recuperação

A branch de trabalho é `recovery/consolidation`. O relatório forense e a lista de inconsistências estão em `ISSUES.md` e `RECOVERY_REPORT.md`.

Antes de qualquer merge em `main`, os workflows atuais devem comprovar build, vet, testes/race, Python adapter e security scan verdes no runner auto-hospedado.

## Princípios

1. Nenhuma funcionalidade existente é removida sem rastreabilidade e justificativa.
2. Contratos, simuladores e adapters não são confundidos com runtimes reais.
3. Falhas e operações não implementadas devem ser explícitas.
4. Segurança é uma fronteira de enforcement, não apenas documentação.
5. Performance só é declarada após benchmark reproduzível.
