# NAME_ONLY_VS_EXECUTABLE

Data da auditoria: 2026-09-29
Repositório de referência do produto: divibisoul/Orquestrador-
Base MAIN inspecionada: default branch, commit a84508b930b691fad9d900d1761263f090df4d17

## 0. Escopo e estado

Esta auditoria não reaudita nem modifica a cadeia RGO → Trinity → MMD → HortaCore.
As integrações das PRs abaixo permanecem preservadas e não são mescladas nesta rodada:

- SARA #24 — rgo-trinity-mmd-2026-09-29-final3 — 650783674951757998e2503eb2d4dec4bdd4ebc9
- N01 #60 — HortaCore bridge — 5bfb9271d9f227176cef8b60cb434435bcc67115
- N07 #54 — rgo-trinity-mmd-2026-09-29-final2 — 98e76f189d3ea217079baf9c5f4a004167fbd586

Estado aceito: cadeia RGO/Trinity/MMD implementada e CI-validada nas três branches; transação distribuída RGO → SARA → N07 → N01 ainda não é LIVE VERIFIED; MAIN ainda não contém a cadeia.

### Legenda

- EXECUTABLE — existe handler real alcançável pelo runtime no ref auditado.
- NAME_ONLY — há anúncio/manifesto/nome no MAIN, mas não há handler efetivo naquele estado.
- SIMULATED — existe código executável de simulação/modelagem, mas ele não representa execução física/produção equivalente.
- BRANCH_ONLY — implementação real existe na branch da PR, mas não está alcançável pelo MAIN.
- BLOCKED_ENV — o código existe, porém a operação fica indisponível sem configuração externa obrigatória.

## A. CAMINHO RGO/TRINITY — BRANCHES DAS PRs 24/60/54

### A.1 SARA #24 — execução/composição real

Entrada HTTP
- src/sara/service/http_api.py
  - POST /v1/rgo/trinity
  - chama system.components["trinity_rgo"].process(...)
  - resultado retorna por as_dict()
  - EXECUTABLE_ON_BRANCH

Composição
- src/sara/rgo/trinity_processor.py
  - RGOTrinityProcessor
  - não implementa uma segunda ARA/ITR/ETR/ERU
  - delega a execução da Tríade à instância existente de TrinityERUUnified
  - vincula as observações à cadeia RGO e ao MMD existente
  - EXECUTABLE_ON_BRANCH

RGO/evidência
- src/sara/rgo/engine.py
  - prepare()
  - ingest_envelope()
  - record_stage_evidence()
  - stage_evidence_integrity()
  - mantém a cadeia de evidência própria dos estágios
  - EXECUTABLE_ON_BRANCH

ERU
- src/sara/meta/eru_engine.py
  - snapshot_state() expõe cópia do snapshot já congelado
  - EXECUTABLE_ON_BRANCH
- src/sara/meta/eru_trinity_bridge.py
  - observação histórica real da Tríade
  - nomes de snapshot incluem digest do estado, evitando colisão entre observações repetidas
  - EXECUTABLE_ON_BRANCH

MMD
- src/sara/omega/soul_services.py
  - reutiliza MicroMacroManager já pertencente ao SoulETROmegaSystem
  - usa transition_explicit(..., evidence=...)
  - EXECUTABLE_ON_BRANCH

Bootstrap
- src/sara/bootstrap.py
  - reutiliza as instâncias canônicas de TrinityERUUnified, ERUTrinityBridge, MicroMacroManager, VagusNerveBus
  - registra rgo e trinity_rgo
  - EXECUTABLE_ON_BRANCH

Conclusão A.1
- A Tríade permanece com uma única autoridade de execução: TrinityERUUnified.
- RGOTrinityProcessor é somente composição/ligação/evidência.
- A herança é de conteúdo/estado/evidência: sequence_index + parent_hash + input_hash + output_hash, não herança de classes.

### A.2 N01 #60 — HortaCore/Soul Mesh

- src/core/mesh/SoulMeshCapabilities.ts
  - anuncia rgo.hortacore.store como capacidade de N01
  - EXECUTABLE_ON_BRANCH
- src/core/mesh/SoulMeshRuntime.ts
  - registra rgo.hortacore.store
  - o handler chama o store real de HortaCore
  - EXECUTABLE_ON_BRANCH
- src/core/rgo/RgoHortaCore.ts
  - grava no HortaCore existente
  - chave inclui finding_id, cycle_id, stage e output_hash, evitando sobrescrita silenciosa de variantes históricas
  - EXECUTABLE_ON_BRANCH
- src/core/rgo/RgoHortaCore.test.ts
  - valida persistência no HortaCore existente
  - TEST CODE ONLY — não é runtime

### A.3 N07 #54 — transporte, correlação e fail-closed

- backend/rgo_proxy.go
  - SARAProxy.RGOTrinity(...)
  - POST real para /v1/rgo/trinity
  - EXECUTABLE_ON_BRANCH
- rgo/trinity_operation.go
  - registra rgo.trinity.process@1.0.0
  - envia o envelope ao SARA
  - recebe stages
  - encaminha cada estágio para N01 via o transportador Mesh existente usando rgo.hortacore.store
  - só retorna promoção de integração quando:
    1. final_status == VALIDATED
    2. toda persistência HortaCore é confirmada
  - caso contrário, retorna estado bloqueado
  - EXECUTABLE_ON_BRANCH
- cmd/nexus/main.go
  - registra rgo.RegisterTrinityOperation(e, saraProxy, peerClient)
  - EXECUTABLE_ON_BRANCH

Conclusão A.3
- Não há segunda Tríade implementada em N07.
- N07 transporta/orquestra; SARA executa a Tríade; N01 é proprietário da memória HortaCore.
- O Soul Mesh continua sendo o transporte inter-núcleos; Vagus não o substitui.

## B. PRODUTO NO MAIN — O QUE O USUÁRIO CLONA HOJE

### B.1 Rotas principais

/v1/health
- anúncio/rota: backend/backend.go
- montagem HTTP: cmd/nexus/main.go via mux.Handle("/v1/", unified.Handler())
- handler real: Server.health
- chama Engine.Health()
- EXECUTABLE
- BLOCKED_ENV para acesso autenticado quando N07_APP_TOKEN não está configurado

/v1/capabilities
- backend/backend.go
- handler real: Server.capabilities
- retorna Engine.Operations() e metadados de SARA/Jev/storage
- EXECUTABLE
- BLOCKED_ENV para acesso autenticado quando N07_APP_TOKEN não está configurado

/v1/execute
- backend/backend.go
- handler real: Server.execute
- valida JSON, cria contexto com timeout, propaga correlation ID e chama Engine.Execute()
- EXECUTABLE
- BLOCKED_ENV para acesso autenticado quando N07_APP_TOKEN não está configurado

/v1/intent
- backend/backend.go
- handler real: Server.intent
- converte ferramentas suportadas em operações do Engine
- EXECUTABLE
- BLOCKED_ENV para acesso autenticado quando N07_APP_TOKEN não está configurado

### B.2 Operações efetivamente registradas no MAIN

O Engine.Operations() só lista registros com handler não nulo. No MAIN, as famílias abaixo são registradas pelo startup real.

Built-ins sempre registrados
- neural.forward@1.0.0
- neural.learn@1.0.0
- compute.execute@1.0.0
- cognitive.execute@1.0.0
- origem: orchestrator/engine.go
- EXECUTABLE

SuperGPU
- supergpu.describe@1.0.0
- supergpu.execute@1.0.0
- supergpu.parallel@1.0.0
- supergpu.memory@1.0.0
- origem: orchestrator/supergpu_ops.go
- EXECUTABLE
- execução continua dependente dos dispositivos/backend efetivamente descobertos; o código não cria GPU física sintética

Advanced
- prefrontal.admission@1.0.0
- mesh.fusion.describe@1.0.0
- mesh.fusion.execute@1.0.0
- supergpu.federated.execute@1.0.0
- origem: orchestrator/advanced_ops.go
- EXECUTABLE

Octacore
- octacore.describe@1.0.0
- octacore.health@1.0.0
- octacore.submit@1.0.0
- octacore.batch@1.0.0
- octacore.signal@1.0.0
- origem: octacore/operations.go
- EXECUTABLE

Jev — condicional
- jev.systemone@1.0.0
- handler: orchestrator/jev.go
- registro no startup somente quando jev.NewFromEnv() é bem-sucedido
- EXECUTABLE WHEN CONFIGURED
- BLOCKED_ENV sem configuração externa necessária

SARA — condicional
- sara.cycle@1.0.0
- sara.audit@1.0.0
- sara.regenerate@1.0.0
- sara.state@1.0.0
- sara.capabilities@1.0.0
- sara.trace@1.0.0
- handlers: backend/sara_proxy.go / RegisterSARAOperations
- registro no startup somente quando SARA_SERVICE_URL e SARA_SERVICE_TOKEN estão configurados
- EXECUTABLE WHEN CONFIGURED
- BLOCKED_ENV sem SARA configurado

### B.3 NAME_ONLY no MAIN

#### Prioridade 1 — SARA anunciado, mas handler não fica registrado sem ambiente

No MAIN, backend/backend.go sempre inclui os seis nomes na seção sara.operations do payload de /v1/capabilities. Porém cmd/nexus/main.go só executa RegisterSARAOperations dentro de if saraProxy.Configured().

Portanto, no MAIN não configurado, os nomes abaixo são NAME_ONLY + BLOCKED_ENV:
- sara.cycle@1.0.0
- sara.audit@1.0.0
- sara.regenerate@1.0.0
- sara.state@1.0.0
- sara.capabilities@1.0.0
- sara.trace@1.0.0

Anunciado em: backend/backend.go, Server.capabilities, seção sara.operations.
Handler efetivo: existe em backend/sara_proxy.go, porém fica ausente do Engine quando SARAProxy.Configured() é falso.

Essa é uma diferença factual entre metadado declarado e executabilidade efetiva; o campo operations do Engine não contém uma operação sem handler.

#### Prioridade 2 — Jev anunciado estaticamente na topologia

orchestrator/topology.go inclui jev.systemone@1.0.0 na lista estática ops e no decision_layer. Entretanto cmd/nexus/main.go só registra o handler por orchestrator.RegisterJevOperations quando jev.NewFromEnv() funciona.

Assim, sem configuração, essa entrada da topologia é NAME_ONLY + BLOCKED_ENV como descrição estática.

Anunciado em: orchestrator/topology.go
Handler efetivo: orchestrator/jev.go
Descoberta runtime: /v1/capabilities deixa jev.operations vazio quando não configurado.

A documentação atual já explicita essa diferença em docs/JEV_INTEGRATION.md.

### B.4 SIMULATED no MAIN

Transcendental Compute Engine (TCE)
- implementação localizada em compute/transcendental/
- executor explícito: compute/transcendental/executor/executor.go → SimulatedExecutor
- documentação: docs/EVOLUTION_AND_COORDINATION.md e compute/transcendental/README.md
- função: referência/simulação para custo, métricas e modelagem de acelerador
- não existe registro público de uma operação TCE equivalente em Engine.registerBuiltins()
- classificação: SIMULATED, não “GPU física” e não “LIVE hardware execution” no MAIN

## C. GAP PARA “USUÁRIO BAIXAR E USAR”

### C.1 Quick Start mínimo do MAIN

Para o usuário sair de clone → subida → health → execute real, os requisitos factuais são:

1. go build ./... / go run ./cmd/nexus no ambiente Go suportado pelo repositório.
2. configurar obrigatoriamente N07_APP_TOKEN para acessar /v1/* autenticados.
3. definir N07_HTTP_ADDR somente se não quiser o padrão :8080.
4. iniciar N07.
5. chamar GET /v1/health com Authorization: Bearer <N07_APP_TOKEN>.
6. chamar GET /v1/capabilities para descobrir as operações efetivamente registradas.
7. executar uma operação local já registrada; um smoke mínimo existente no próprio repositório usa neural.forward@1.0.0 com vetor de 8 valores.

Os serviços externos não são necessários para o primeiro smoke neural local, mas são necessários para capacidades que dependem deles. Em particular:
- SARA: SARA_SERVICE_URL + SARA_SERVICE_TOKEN
- Jev: JEV_API_KEY (e configuração de base URL conforme integração)
- armazenamento: credenciais de Supabase/Web3 Storage quando essas funções forem usadas

### C.2 O que ainda não fica no MAIN

A cadeia RGO/Trinity/MMD não está no MAIN porque as alterações permanecem nas PRs/branches congeladas.

Para expô-la após merge ordenado será necessário, no mínimo:
- SARA #24 alcançável no tronco padrão, inclusive /v1/rgo/trinity
- N01 #60 alcançável no tronco padrão, inclusive rgo.hortacore.store
- N07 #54 alcançável no tronco padrão, inclusive rgo.trinity.process@1.0.0
- configuração real de URLs/tokens dos serviços
- transação distribuída correlacionada observada ponta a ponta
- comprovação de persistência em HortaCore

A simples presença dos arquivos após merge não será tratada como LIVE VERIFIED. LIVE VERIFIED exige requisição real com URLs reais, correlação observável e respostas verificadas em cada fronteira.

### C.3 Ordem de integração proposta — não executada

Ordem lógica de menor risco para expor a cadeia, sem executar nesta rodada:

SARA #24 → N01 #60 → N07 #54

Depois do merge ordenado, o smoke deve validar:

client → N07 /v1/execute → rgo.trinity.process → SARA /v1/rgo/trinity → N01 rgo.hortacore.store

com um único correlationId e evidência de final_status=VALIDATED + persistência de todas as etapas.

## D. VEREDITO DA AUDITORIA

- PRESERVADO: PRs 24/60/54, TrinityERUUnified, RGOTrinityProcessor como composição, chain de hashes, MMD/Vagus/HortaCore existentes, Mesh 1.1.0 e fail-closed.
- MAIN: possui um núcleo operacional real para health/capabilities/execute e um conjunto real de handlers; não é apenas nome.
- NAME_ONLY no MAIN: principalmente as operações SARA declaradas na seção de capabilities quando SARA não está configurado e a presença estática de Jev na topologia sem serviço configurado.
- SIMULATED: TCE isolado, explicitamente modelado como simulador/referência.
- BRANCH_ONLY: RGO/Trinity/MMD/HortaCore bridge desta rodada.
- LIVE: não verificado; nenhum claim de produção foi criado por esta auditoria.

Nenhuma feature nova foi implementada.
Nenhuma das PRs 24/60/54 foi mesclada.
