# NAME_ONLY_VS_EXECUTABLE — FASE 0 GLOBAL

Data: 2026-09-30
MAIN N07 observado: abf01b43611a582bf0205be49818c5925fccde0f

## Estado aceito das PRs congeladas

A cadeia RGO/Trinity/MMD/HortaCore permanece congelada e preservada:
- SARA #24
- N01 #60
- N07 #54

Nenhuma dessas três PRs é mesclada neste documento.

## MAIN — diferença entre nome e código executável

### EXECUTABLE
O MAIN do Orquestrador possui handlers reais para:
- neural.forward@1.0.0
- neural.learn@1.0.0
- compute.execute@1.0.0
- cognitive.execute@1.0.0
- supergpu.describe@1.0.0
- supergpu.execute@1.0.0
- supergpu.parallel@1.0.0
- supergpu.memory@1.0.0
- octacore.describe@1.0.0
- octacore.health@1.0.0
- octacore.submit@1.0.0
- octacore.batch@1.0.0
- octacore.signal@1.0.0

Rotas principais reais:
- GET /health
- GET /v1/health
- GET /v1/capabilities
- POST /v1/execute
- POST /v1/intent
- GET /v1/models
- POST /v1/chat/completions
- GET /metrics
- GET /identity
- GET /topology
- POST /api/soul-mesh

## BLOCKED_ENV
SARA:
- SARA_SERVICE_URL
- SARA_SERVICE_TOKEN

Jev:
- JEV_API_KEY

Peers Mesh:
- URLs/tokens/HMAC conforme peer

Storage:
- Supabase/Web3 credentials

## NAME_ONLY
### SARA
backend/backend.go anuncia seis operações na seção metadata sara.operations:
sara.cycle@1.0.0
sara.audit@1.0.0
sara.regenerate@1.0.0
sara.state@1.0.0
sara.capabilities@1.0.0
sara.trace@1.0.0

Os handlers existem em backend/sara_proxy.go, mas o startup registra a família somente quando SARAProxy.Configured() é verdadeiro. Sem configuração, o nome aparece no metadata mas não existe no Engine registry.

### Jev
orchestrator/topology.go descreve jev.systemone@1.0.0 estaticamente. orchestrator/jev.go contém o handler, mas o startup só o registra quando JEV_API_KEY permite criar o cliente.

## BRANCH_ONLY
RGO/Trinity/MMD/HortaCore:
- SARA #24
- N01 #60
- N07 #54

Fase 0 N01 Core continuity: PR #70, head 91e311924aa047af692b6701c60745d80e0ce08e. CI final do guard e validações principais do N01: SUCCESS. PR ainda OPEN; não mergeada.

Até merge ordenado, esses handlers não fazem parte do MAIN.

## SIMULATED
TCE:
compute/transcendental/executor/executor.go contém SimulatedExecutor. É simulação/modelagem; não é execução de GPU física nem LIVE hardware.

## Ordem de Merge — proposta, não executada

1. SARA #24
   - disponibiliza /v1/rgo/trinity e composição sobre TrinityERUUnified.
2. N01 #60
   - disponibiliza rgo.hortacore.store no HortaCore existente.
3. N07 #54
   - disponibiliza rgo.trinity.process@1.0.0 e fail-closed.

Depois de cada merge, executar o gate da etapa correspondente. Só após os três merges:
client → N07 /v1/execute → rgo.trinity.process → SARA /v1/rgo/trinity → N01 rgo.hortacore.store.

Critério de LIVE VERIFIED:
- URLs reais;
- credenciais reais;
- um correlationId preservado;
- respostas observadas em cada fronteira;
- final_status=VALIDATED;
- persistência HortaCore de todas as etapas comprovada.

A mera existência de arquivos, PRs ou CI não cria evidência LIVE.

## Coordenação com frentes paralelas

Frentes ativas devem continuar como unidades independentes e ser integradas por dependência, não por duplicação:
- N01: recovery/Mesh/Horta/telemetria/Gemini/ATLAS
- N02: Gemini primordial tools/ATLAS/capability authority
- N03: real Gemini chat/correlation/Mesh
- N04: tool compatibility/ATLAS/Mesh
- N05: real inference/ATLAS/Mesh
- N06: context/ATLAS/Mesh
- N07: NeuralForge/ASC/cognitive/primordial/ATLAS
- SARA: Bayesian/OctaCore/Vagus/RGO governance

Nenhuma dessas frentes é tratada como substituta de outra.

## Regra de decisão

DECLARED ≠ EXECUTABLE ≠ RUNNING ≠ VERIFIED ≠ LIVE VERIFIED.
