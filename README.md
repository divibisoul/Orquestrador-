# ORQUESTRADOR-NEXUS

Plataforma cognitiva distribuída e provider-neutral. Esta recuperação consolida a fundação existente sem apagar funcionalidades por suposição.

## Arquitetura

```text
Objetivo
   |
   v
Neo-Córtex -> Neural Fabric -> Orquestrador -> Compute
   |                              |
   +------------------------------+
                  |
               Mesh <-> State
                  |
             API / Protobuf
```

Pacotes canônicos de domínio: `core/orchestrator`, `core/prefrontal`, `core/neuralfabric`, `core/superagi`. `mesh`, `state`, `compute` e `security` fornecem infraestrutura/runtime. Pacotes `internal/*` permanecem como compatibilidade até que seus consumidores sejam migrados e sua remoção seja comprovadamente segura.

## Maturidade

- **Implementado:** workflow/DAG e primitives de resiliência; planejamento/decisão do Neo-Córtex; roteamento básico; registro/descoberta/heartbeat do Mesh; estado versionado/cache; políticas básicas de segurança.
- **Parcial/boundary:** Neural Fabric adaptativa/persistência; Super AGI com provedores/modelos reais; GPU/NPU; Raft distribuído; mTLS/SPIFFE; observabilidade externa; gRPC runtime.
- **Experimental:** adapter Gemini da branch `feature/gemini-provider`; não é considerado parte segura do runtime canônico até revisão específica.

## Executar

Requisitos: Go conforme `go.mod`.

```bash
go mod tidy
go build ./...
go test ./...
go test ./... -race
go run ./cmd/nexus
```

O plano de controle Nexus usa `:8080` e fornece os endpoints documentados em `docs/API_REFERENCE.md`.

## Recuperação

A branch de trabalho é `recovery/consolidation`. O relatório forense e a lista de inconsistências estão em `ISSUES.md` e `RECOVERY_REPORT.md`.

Antes de qualquer merge em `main`, os workflows do GitHub Actions devem comprovar build, vet, testes e security scan verdes.

## Princípios

1. Nenhuma funcionalidade existente é removida sem rastreabilidade e justificativa.
2. Contratos e adapters não são confundidos com runtimes reais.
3. Falhas e operações não implementadas devem ser explícitas.
4. Segurança deve ser uma fronteira de enforcement, não apenas documentação.
5. Performance só é declarada após benchmark reproduzível.
