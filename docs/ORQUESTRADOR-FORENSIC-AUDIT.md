# Orquestrador — auditoria forense e mapa de prontidão

## Estado auditado

O repositório atualmente contém uma implementação funcional isolada do Transcendental Compute Engine (TCE), mas ainda não contém runtime.go, router.go ou cortex.go. Portanto, a integração com Prefrontal/Neural Fabric não pode ser declarada existente.

## Áreas

| Área | Estado | Diagnóstico | Ação |
|---|---|---|---|
| Core/contratos | forte | structs, validação e contexto presentes | endurecer invariantes |
| Modelos | forte | cinco modelos e catálogo canônico | manter referências versionadas |
| Selector | forte | memória, precisão, FP64 e modos | ampliar scoring/benchmarks |
| Executor | forte | execução, plano, histórico e teto de workers | melhorar telemetria e erros por item |
| Interfaces | forte | contratos para backend/metrics/cost | preservar desacoplamento |
| Configuração | média | Config existe e feature flag off | adicionar loader externo sem contaminar pacote |
| Testes | média | unit/integration existentes | adicionar fuzz, race e propriedades |
| CI | forte | format, vet, test, race, build, govulncheck | manter como gate |
| Observabilidade | média | histórico local de Metrics | preparar health/capability registry |
| Segurança | média | sem rede/driver no TCE | manter isolamento e scan de dependências |
| Integração Orquestrador | ausente por desenho atual | não existem os arquivos-alvo | somente após leitura dos consumidores reais |

## Regras de anticorpo

- Nunca declarar integração por presença de arquivos.
- Nunca declarar E2E sem execução real.
- Rejeitar workload inválido antes de seleção/execução.
- Rejeitar modelo com PFLOPS/bandwidth inválidos.
- Preservar feature flag disabled como caminho seguro.
- Não adicionar transporte de rede ao TCE isolado.
- Toda alteração deve ser seguida por leitura do arquivo no GitHub e CI.

## Preparação para a peça faltante

A futura camada de integração deve consumir apenas:

- `interfaces.ComputeBackend`
- `interfaces.MetricsProvider`
- `interfaces.CostEstimator`

O consumidor futuro deve fornecer adapters opcionais e nil-safe. O TCE não deve importar Prefrontal, Neural Fabric, router ou cortex.

## Ordem de evolução

1. Core e contratos
2. Catálogo e modelos
3. Selector
4. Executor
5. Histórico/telemetria
6. Configuração externa
7. Testes unitários + propriedades + fuzz
8. CI/security
9. Adapter para Prefrontal
10. Adapter para Neural Fabric
11. Integração com Orquestrador real
12. E2E entre peças

## Nota sobre referências de hardware

Os números do modelo são referências matemáticas e não representam hardware real. Vera Rubin e Atlas 950 permanecem explicitamente preliminares/scale-out até existirem especificações oficiais adequadas ao modelo.
