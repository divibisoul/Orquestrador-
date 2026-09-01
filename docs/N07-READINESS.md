# N07 — Readiness e Matriz das 40 Funções

> Regra: a função só é considerada pronta quando implementação e comportamento têm evidência executável. Integração externa é uma dimensão separada e não é inferida pela existência do código.

## Estado da frente

Branch: `upgrade/n07-production-v7`

Último commit verificado nesta rodada: `3e08c01b1fb3552d280ae60460c14ec79821cfa7`

### Gráfico operacional

```text
Implementação das 40 funções    ████████████████████ 100% declaradas/implementadas
Evidência executável             ████████████████████ 40/40 por teste direto/fluxo
Integração Mesh/N07              ███████████████░░░░░ 75% contrato + ponte executável
N01–N06 externo real             █████░░░░░░░░░░░░░░ 25% ainda não demonstrado
Ferramentas/funções fusionadas   █████░░░░░░░░░░░░░  25% contrato preparado; fusão final pendente
Produção/activation gate         █████░░░░░░░░░░░░░  25% dependente das provas externas
```

## Matriz por núcleo

| Núcleo | Funções | Implementação | Evidência executável | Integração |
|---|---:|---:|---:|---:|
| Orchestrator | 10 | 10/10 | 10/10 | 8/10 |
| Neural | 10 | 10/10 | 10/10 | 6/10 |
| Prefrontal | 10 | 10/10 | 10/10 | 7/10 |
| SuperGPU | 10 | 10/10 | 10/10 | 6/10 |
| **Total** | **40** | **40/40** | **40/40** | **27/40** |

## Pontes que já existem

```text
SOUL Mesh envelope
      |
      v
MessageFromMesh
      |
      v
N07 protocol
      |
      v
Engine.SubmitMesh
      |
      v
registered handler / builtin pipeline
      |
      v
MeshFromMessage
      |
      v
SOUL Mesh response
```

A ponte preserva `correlationId/trace_id`, origem/destino, payload numérico legado e payload JSON genérico.

## Gaps que continuam ativos

- demonstração contra endpoints reais N01–N06;
- readiness do Mesh participando diretamente da seleção/fallback distribuído;
- exportação de tracing para backend externo;
- identidade SPIFFE/workload real;
- backends CUDA/ROCm/NPU executáveis;
- LoRA de treino/execução real;
- Raft integrado ao scheduler/recovery distribuído;
- fusão final de ferramentas e funções N07/N01/N06;
- activation gate de produção.

## Regra de evolução

N07 continua sendo desenvolvido agora. A fusão final acontece somente depois que as entradas/saídas, eventos, ferramentas, funções e capacidades de N01–N06 forem mapeados e os contratos compatíveis forem comprovados.
