# SOUL — Fase 0 Composition Coverage Ledger — 2026-09-29

## Purpose

Este ledger mede o **encaixe estrutural das composições já existentes**. Ele não transforma CI, HTTP 200 ou catálogo declarativo em prova de execução online.

A métrica abaixo foi reconstruída imediatamente antes da correção desta frente a partir do código atual de `main`. Ela não é uma porcentagem histórica previamente registrada; é o **baseline reproduzível desta auditoria**.

## Baseline reconstruído antes da correção

| Link composto | Gates avaliados | Gates atendidos | Cobertura |
|---|---:|---:|---:|
| N03 → N05 (percepção → inferência) | 5 | 3 | 60% |
| N05 → N04 (inferência → ferramenta) | 5 | 3 | 60% |
| N06 → N05/N04 (cadeias canônicas) | 5 | 3 | 60% |
| **Média desta frente** | **15** | **9** | **60%** |

### Gates

1. capability ID existe no catálogo/registro do destino;
2. handler/runtime correspondente existe;
3. composição usa a autoridade atual da capability;
4. transporte/correlação da Soul Mesh está disponível;
5. payload/contrato não depende de uma capability inexistente.

O baseline de 60% falhava principalmente por **desalinhamento de IDs e ownership efetivo**, não por ausência geral de infraestrutura.

## Correções executadas nesta frente

### N03
Branch: `fix/phase0-composition-contract-2026-09-29`

- `perceptionToReasoning()` passa a usar N05 como proprietário atual de `inference.reason`.
- `perceptionToExecution()` e `perceptionReasoningExecution()` deixam de fabricar um payload de ferramenta genérico.
- Um payload de ferramenta executável passou a ser obrigatório.
- Testes adicionados para travar N03 → N05 e N03 → N04 contra regressão.

### N04
Branch: `fix/phase0-tool-capability-alias-2026-09-29`

- `tool.execute` foi adicionado como **alias compatível e não destrutivo** do boundary `tool-execution`.
- O alias foi adicionado ao catálogo, processor e Mesh runtime.
- O comportamento existente `tool-execution` permanece intacto.
- Testes cobrem descoberta e execução do alias.

### N05
Branch: `fix/phase0-inference-capability-catalog-2026-09-29`

- O catálogo Mesh passou a publicar as capacidades que o runtime já implementava:
  - `inference.reason`
  - `inference.analyze`
  - `inference.summarize`
  - `inference.translate`
  - `inference.classify`
  - `conversation.chat`
  - `conversation.memory`
- Testes garantem que catálogo e gateway possuem o mesmo conjunto executável.

### N06
Branch: `fix/phase0-composition-targets-2026-09-29`

- Cadeia cognitiva foi alinhada para N05 → N04 → N05.
- `inference.reason` deixou de ser delegado a N02 nessas cadeias.
- `conversation.summarize` foi substituído por `inference.summarize`, capability publicada pelo N05.
- A cadeia de composição agora exige payload executável explícito para ferramenta/documento em vez de gerar instruções fictícias.
- O handoff N06 → N01 também foi corrigido para encaminhar `inference.reason` ao N05.

## Estado depois da correção

Para os gates estruturais desta frente:

| Link composto | Gates atendidos | Cobertura estrutural |
|---|---:|---:|
| N03 → N05 | 5/5 | **100%** |
| N05 → N04 | 5/5 | **100%** |
| N06 → N05/N04 | 5/5 | **100%** |
| **Média desta frente** | **15/15** | **100% estrutural** |

### O que 100% NÃO significa

Ainda não significa que os três processos estão simultaneamente online.

A prova E2E continua separada:

`DISCOVER → NEGOTIATE → EXECUTE → CORRELATE → VERIFY`

Somente um teste real entre runtimes implantados pode promover a composição para `VERIFIED`.

## Regra de continuidade

Esta correção não encerra a auditoria. Enquanto estes PRs são validados, a próxima frente deve simultaneamente:

- continuar a arqueologia de módulos soterrados;
- reconciliar os demais IDs/ownerships divergentes;
- verificar N02/N03/N04/N05/N06/N07 e SARA contra os mesmos contratos;
- atualizar este ledger com novos deltas;
- nunca substituir uma capacidade existente por um proxy apenas para satisfazer a métrica.

## Evidência desta execução

- alterações reais realizadas em branches GitHub dedicadas;
- nenhuma exclusão de módulo;
- nenhum mock criado para representar execução real;
- nenhum novo Mesh criado;
- percentual calculado a partir de gates explícitos e reproduzíveis.

## Pendência transversal preservada

A cobertura de 100% acima é restrita aos **links de composição reparados nesta frente**. A auditoria global ainda encontrou manifests `SoulOwnership.yaml` históricos em múltiplos repositórios que divergem da autoridade consolidada em `N05OwnershipMatrix.ts`. Esses manifests não foram apagados nem sobrescritos nesta frente para não misturar migração de governança com correção de execução.

Estado da pendência: **OPEN / GOVERNANCE-RECONCILIATION**.

Próxima correção: comparar cada capability publicada com seu proprietário efetivamente executável e migrar os manifests legados de forma aditiva, preservando histórico e compatibilidade.

## Mainline recovery state — verified 2026-09-29

| Center | Mainline change now present | Exact main SHA | Status |
|---|---|---|---|
| N01 | Aeternum buried module recovery + HortaCore/EventBus continuity | `15b40a206caf9fa2012e68d94f7d32ecc68bd65f` | MERGED |
| N02 | BNCv2 + CSAE + DCRS + VagusBus recovery | `4f95c50b2841829daf2904fdb7c7a8171e1a9113` | MERGED |
| N03 | N03→N05→N04 composition contract repair | `9951485529c8480ff3bf4d7305538d152b6acdfe` | MERGED |
| N04 | `tool.execute` compatibility boundary | `ed4e2a75b772d1efaed2457c392f52bddaffbfa8` | MERGED |
| N05 | executable inference/conversation capability catalog | `e0c8778bebaba4a4b821bc6923c2a51dcfc431ca` | MERGED |
| N06 | N05→N04 composition target repair | `bdcb099166701453b6cb62672fa6a642addbd9b0` | MERGED |
| N07 | existing neural/prefrontal runtime verified on main; recovery PR was test-only | `3516c74c52da2b7d423290c92350787a788a341d` | RUNTIME ALREADY PRESENT |
| SARA | Chimera + OctaCore/Vagus/Mesh recovery | `50c9d0f12a42f518994eba8ee99ea1c40035b38a` | MERGED |

**Important:** this table records repository state, not online runtime availability. E2E commissioning remains a separate proof layer.
