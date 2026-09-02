# SOUL — FRENTE 3 REPORT

Data: 2026-09-02
Branch: `frente-3`

## Revisão antes da execução

A revisão do estado real mostrou que o cenário recebido precisava de duas correções de contexto:

- N02 PR #13 (`feat(mesh): resilient N01-N02 canonical gateway`) já estava **merged** em 2026-09-02. Não é tecnicamente possível mantê-lo aberto ou bloqueá-lo retroativamente.
- N07 não possui `.github/workflows/n07-e2e-commissioning.yml`; o workflow operacional existente é `.github/workflows/executor-runner-validation.yml`. As alterações foram aplicadas nele para não criar um caminho paralelo de CI.

O N07 hardening anterior estava em PR #22, aberto e draft. A Frente 3 foi criada a partir do HEAD real desse hardening (`f67357bb503d9541d24e88930e708c1cbbe045d7`) para preservar as alterações existentes e adicionar somente a camada incremental desta frente.

## Alterações efetivadas

### N02

Criado `.github/workflows/merge-gate.yml` na branch `frente-3` do N02.

O gate executa:

- `npm install --no-audit --no-fund`
- `npm run lint`
- `npm run build`

Cada etapa produz outcome no relatório `.ci/validation-report.json`, que é publicado como artefato. Qualquer falha mantém o job em estado FAILED; não há `continue-on-error`, `process.exit(0)` artificial ou supressão de erro.

A workflow também mantém a etiqueta `ci-pending` enquanto a validação de um PR não estiver bem-sucedida. A exigência de bloqueio efetivo de merge precisa ser configurada como check obrigatório nas regras de proteção/ruleset de `main`; o workflow não pode bloquear um PR histórico já mesclado.

### N07 — Secret governance

Criado `scripts/pre-sync-check.js`.

O pré-check:

- não lê nem imprime valores de secrets;
- grava presença/ausência em `.diagnostics/pre-sync-status.json`;
- falha explicitamente com `BLOCKED: Missing required secrets...` quando `SOUL_MESH_TOKEN` ou `REPO_SECRETS_ADMIN_TOKEN` está ausente.

`sync-secrets.yml` agora chama esse pré-check antes de instalar dependências ou tentar sincronizar qualquer secret.

### N07 — Dashboard

`api/health/dashboard.go` foi endurecido sem remover a autenticação existente por `N07_APP_TOKEN`.

Quando `SOUL_MESH_PEERS` não existe, o inventário ainda contém N01–N06 e cada probe retorna `not-configured`, com:

- `CureAction = configure_peers`
- sugestão para configurar `SOUL_MESH_PEERS` com URLs reais.

Quando todos os checks estão `not-configured`, `overallStatus = not-configured`.

As sondagens continuam persistidas no histórico. A tendência ignora amostras `not-configured` e começa a calcular evolução somente depois de haver pelo menos duas amostras ativas com latência válida.

### N07 — E2E

`executor-runner-validation.yml` agora tem uma etapa de descoberta obrigatória e uma etapa de execução verificável.

Antes de executar o teste nominal, o workflow exige que:

`TestN01ToN07FederatesAcrossN04N05N06`

seja listado por `go test ./mesh -list ...`. Se o teste não for encontrado, o job falha explicitamente.

A execução usa `-count=1 -v`, timeout de 180 segundos, e exige os marcadores `=== RUN` e `--- PASS` do teste exato. Assim, `PASS [no tests to run]` não pode ser aceito como evidência de certificação.

O log da execução é arquivado para auditoria. Os 30 probes dirigidos (15 pares × 2 direções) continuam sendo executados com `if: always()` após a etapa nominal e geram relatório próprio.

## O que permanece bloqueado por configuração real

- `SOUL_MESH_TOKEN` ainda precisa existir como secret real no N07.
- `REPO_SECRETS_ADMIN_TOKEN` ainda precisa existir no N07 com escopo mínimo de Actions secrets write nos N01–N06.
- `SOUL_MESH_N01_URL` … `SOUL_MESH_N06_URL` ainda precisam apontar para endpoints reais para que os 30 probes tenham valor operacional.
- `N07_APP_TOKEN` precisa existir para acesso autenticado ao dashboard.
- A certificação E2E final ainda depende da execução real do workflow após estas mudanças.

## Evidência e limites

Implementação não é confundida com certificação. Enquanto os secrets/URLs não existirem, a condição correta é `not-configured`, nunca `healthy`.

Os 30 probes do N07 comprovam o contrato e a rota a partir do executor N07; não são, isoladamente, prova física de que cada processo remoto originou a chamada. Essa prova continua dependente da execução do teste federado e da instrumentação real disponível na topologia.

## Estado de CI desta execução

O PR da Frente 3 está aberto e a execução final do workflow ainda está **PENDENTE DE EVIDÊNCIA** neste momento. Nenhum PASS é declarado sem logs/jobs reais.

N04 e N06 não foram modificados.
