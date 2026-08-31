# ORQUESTRADOR-NEXUS

Plataforma cognitiva distribuída e provider-neutral. A linha de finalização consolida a fundação, Trinity, Mesh/SOUL, estado, segurança, transporte e fronteiras de provider em uma base canônica, testável e evolutiva.

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
                  |
           Provider boundaries
```

Pacotes canônicos de domínio:

- `core/orchestrator` — engine único de workflow, execução, resiliência e recovery.
- `core/prefrontal` — planejamento, decisão, conflito, working memory e feedback executivo.
- `core/neuralfabric` — seleção de rota e feedback adaptativo para a Trinity, preservando a Neural Fabric histórica.
- `core/trinity` — contrato e fluxo PFC -> Fabric -> Compute -> Feedback -> memória.
- `core/superagi` — geração, inferência, memória, verificação e lifecycle de aprendizado sobre uma única fonte de verdade.
- `mesh` — registro, descoberta, heartbeat, stale-node e capacidades.
- `core/soul` — integração canônica dos núcleos N01-N06 sem copiar seus runtimes.
- `state` — armazenamento versionado, CAS e durable store.
- `state/raft` — núcleo de consenso distribuído isolado, pronto para integração controlada ao recovery.
- `compute` — abstração de dispositivos e execução local; backends físicos são providers explícitos.
- `mesh/security` — mTLS TLS 1.3 e verificação opcional de identidade SPIFFE.
- `api/grpc` — servidor/cliente gRPC e health checking; HTTP legado permanece disponível.
- `observability` — métricas e propagação de trace context.

## Estado de maturidade

A linha atual está em **hardening/finalização**. O critério de conclusão é comportamento demonstrado, não quantidade de arquivos.

**Funcionando e testado:** workflow lifecycle; execução concorrente/distribuída; retry; circuit breaker; bulkhead; rate limiter; PFC; working memory; Neural Fabric; Trinity; compute CPU determinístico; Mesh registry; SOUL envelope/trace validation; durable store; núcleo Raft; mTLS; gRPC round-trip/health; adapters SuperAGI para generator/inference/learning/memory/verifier; provider Gemini como boundary injetável.

**Ainda dependente de ambiente/provider real:** GPU CUDA/ROCm/NPU físicos; workload identity SPIFFE provisionada por uma autoridade de identidade; treinamento efetivo de pesos LoRA; E2E entre os seis repositórios ativos; integração operacional de Raft ao scheduler/recovery distribuído; tracing/exportador externo; CI de segurança do HEAD final.

Nada desses itens é apresentado como funcional só porque possui interface ou catálogo.

## Execução

Requisitos: Go conforme `go.mod` e Python 3 para os adapters.

```bash
go mod tidy
go build ./...
go test ./...
go test ./... -race
go run ./cmd/nexus
go run ./cmd/orchestrator
```

O plano de controle Nexus usa `:8080`. Consulte `docs/API_REFERENCE.md`.

## Feature flags

A Trinity permanece opt-in. O caminho legado continua intacto quando a integração não está habilitada. A ativação completa deve ocorrer somente depois de todos os gates de validação.

## Princípios

1. Nenhuma funcionalidade existente é removida sem rastreabilidade e justificativa.
2. Uma única implementação canônica por responsabilidade; adapters delegam, não duplicam.
3. Contratos, simuladores e boundaries de provider não são confundidos com runtimes físicos.
4. Toda falha encontrada entra imediatamente no ciclo de correção e teste.
5. Estado compartilhado é protegido contra concorrência, aliasing e recuperação inconsistente.
6. Performance e disponibilidade só são declaradas após teste reproduzível.
7. Branches de outras frentes são tratadas como trabalho do mesmo sistema: antes de modificar, comparar; diante de conflito, integrar a solução mais completa sem apagar a outra.
