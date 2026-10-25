# SOUL — auditoria dos upstreams públicos e equilíbrio de capacidades
**Data:** 2026-10-25  
**Base N07 main:** `c4582f178a6bc82ad7a7a2ebbe4332cd7982439d`  
**Branch de auditoria:** `audit/external-capability-balance-2026-10-25`  
**Escopo:** os 16 upstreams públicos catalogados no N07 e seus contratos de ownership, adapters e acesso pelos núcleos nativos.

## Resumo executivo

- **16/16 upstreams catalogados** no registro `integrations/external-capabilities.json`.
- **16/16 gitlinks/submodules pinados** sob `integrations/external/*`; cada gitlink possui SHA explícito. Isso é proveniência reproduzível, não prova de runtime.
- Há um runner central e operações declaradas para os 16 upstreams. A execução externa é **desligada por padrão** e deve permanecer assim até a prova isolada por provider.
- O validador estrutural confirma que o catálogo, a lista de adapters e a lista de ownership contêm os mesmos 16 IDs e que cada owner pertence aos targets declarados. O teste Go existente confirma registro de 16 providers, fail-closed por padrão e probe de LangGraph.
- **Nenhum dos 16 foi provado como execução externa live autenticada nesta auditoria.** Não promover estado a `adapter-live` apenas por código, submodule ou CI.

## Inventário pinado

| Upstream | Repositório público | SHA fixado | Owner principal | Capacidade-alvo |
|---|---|---|---|---|
| Superpowers | [obra/superpowers](https://github.com/obra/superpowers) | `8ca22dba9a94f28898bbce59f2537ff4d87c747d` | N07 | método de engenharia, skills, TDD/revisão |
| SuperAGI | [TransformerOptimus/SuperAGI](https://github.com/TransformerOptimus/SuperAGI) | `c3c1982e7bd6a11cfed53c5a193ea502f924b1b6` | N07 | agentes, ferramentas e memória |
| LangGraph | [langchain-ai/langgraph](https://github.com/langchain-ai/langgraph) | `157a06dda988d85afeb8751ff27b35ab3f4f8bf4` | N07 | workflow com estado e orquestração |
| CrewAI | [crewAIInc/crewAI](https://github.com/crewAIInc/crewAI) | `8078f9130c35a47be95d4a55bf1d73b3fd44fc88` | N07 | equipes multiagente |
| Microsoft Agent Framework | [microsoft/agent-framework](https://github.com/microsoft/agent-framework) | `a2f4506c0ba30cea7c9bbe907fc158c0db2cc6a3` | N07 | agentes, workflows, MCP/A2A |
| OpenHands | [OpenHands/OpenHands](https://github.com/OpenHands/OpenHands) | `2414d6ee5e31bede2e78211f72b58e9949575a75` | N06 | agente de engenharia de software |
| MetaGPT | [FoundationAgents/MetaGPT](https://github.com/FoundationAgents/MetaGPT) | `11cdf466d042aece04fc6cfd13b28e1a70341b1f` | N07 | papéis e processo de software |
| AgentScope | [agentscope-ai/agentscope](https://github.com/agentscope-ai/agentscope) | `72f3f6fa0b2fc38b8517f408ab616f0f2bd229e6` | N03 | agentes, equipes e voz |
| Letta Code | [letta-ai/letta-code](https://github.com/letta-ai/letta-code) | `1fcc9666817ab852bc2532a3a989f712e1fd6c19` | N01 | memória persistente e identidade |
| Browser Use | [browser-use/browser-use](https://github.com/browser-use/browser-use) | `302d8fcb245a7a63fb7531a4734c9ce3c7792779` | N04 | automação de navegador |
| SmolAgents | [huggingface/smolagents](https://github.com/huggingface/smolagents) | `c30b115286e000e98711fae5e85993547b73d826` | N06 | agentes de código e ferramentas |
| Pydantic AI | [pydantic/pydantic-ai](https://github.com/pydantic/pydantic-ai) | `6bc07cf18b0641ea92343d8c589cfb922108b802` | N01 | agentes tipados e contratos |
| LlamaIndex | [run-llama/llama_index](https://github.com/run-llama/llama_index) | `962940ddc079cc21701d28d1237c84c82a7c5164` | N05 | RAG, indexação e recuperação |
| DSPy | [stanfordnlp/dspy](https://github.com/stanfordnlp/dspy) | `ba3f9198efe5d125c7c1a2b40b1f1e6166209bd2` | N06 | programação/otimização de pipelines de raciocínio |
| Whisper | [openai/whisper](https://github.com/openai/whisper) | `86098128c0b4f24f0e2aa2994de830614b474227` | N03 | transcrição de fala |
| Kokoro | [hexgrad/kokoro](https://github.com/hexgrad/kokoro) | `dfb907a02bba8152ca444717ca5d78747ccb4bec` | N03 | síntese de fala |

## Estado por camada

| Camada | Evidência observada | Estado |
|---|---|---|
| Fonte e proveniência | 16 gitlinks pinados e registrados em `.gitmodules` e `integrations/external-capabilities.json` | **REAL — estrutural** |
| Catálogo de capacidades | 16 IDs e targets por núcleo no registro | **REAL — estrutural** |
| Ownership | owner primário e consumers definidos separadamente | **REAL — estrutural; contrato precisa permanecer canônico** |
| Adapter central | `scripts/external-capability-adapter.py` e operações no manifesto | **REAL — código presente** |
| Registro no executor N07 | `orchestrator/external_adapters.go` e testes de registro | **REAL — código/testes presentes** |
| Provider importado e executado com sucesso | Não executado nesta auditoria para os 16 providers | **PROJECTED / UNVERIFIED** |
| E2E pelo Mesh com owner, HMAC, correlação e resultado validado | Não demonstrado para os 16 providers | **PROJECTED / UNVERIFIED** |

## Achados de equilíbrio e possíveis conflitos

### E1 — Semântica de estado divergente entre documentos

`integrations/external-capability-ownership.json` marca todos os 16 como `adapter-planned`, enquanto `integrations/native-capability-augmentation.json` usa `ADAPTER_BOUND`. Essas expressões não são equivalentes: a primeira diz adapter previsto/estrutural, a segunda pode ser lida como integração pronta. **Não alterar para `adapter-live` sem execução E2E.** Próxima correção recomendada: normalizar os documentos para o vocabulário canônico `structural-only / adapter-planned / adapter-live` e definir `adapter-planned` como padrão até teste de execução.

### E2 — Duas raízes históricas de integração

Além de `integrations/external/*`, o repositório contém raízes como `integrations/langgraph`, `integrations/superpowers` e `integrations/nemo-agent-toolkit`. Elas não são listadas como os mesmos submodules canônicos de `integrations/external/*` no `.gitmodules`. **Não apagar nem fundir por suposição.** É necessário identificar se são protótipos, adapters locais, assets ou uma segunda implementação antes de decidir seu papel. Até essa reconciliação, não criar uma segunda rota de execução para o mesmo provider.

### E3 — CI de contrato manual apenas

`.github/workflows/external-adapter-contract.yml` e `.github/workflows/external-ownership-resident-contract.yml` declaram somente `workflow_dispatch`. Isso significa que uma alteração normal de código pode não executar esses gates automaticamente. Recomenda-se criar gatilhos de PR/push filtrados para os manifestos, scripts, adapters, testes e contratos relevantes; isso deve ser feito sem reativar fan-out de workflows irrelevantes.

### E4 — Registro não é execução

Os testes existentes demonstram registro e fail-closed de LangGraph, além do probe. Eles não demonstram que todos os 16 providers importam, iniciam, chamam modelo/serviço e retornam resultado. Cada provider precisa de dois níveis separados: teste determinístico de contrato sem credenciais e smoke/E2E opcional com dependências reais.

### E5 — Dependências, licenças e segurança upstream

O inventário atual fixa revisões, mas esta rodada não completou uma matriz de licenças por arquivo/pacote, vulnerabilidades transitivas, versões de runtime, consumo de memória/CPU, compatibilidade de dependências nem revisão de permissões. Isso precisa ser verificado antes de ativar qualquer provider. O pinning reduz deriva inesperada, mas não prova segurança nem compatibilidade.

### E6 — Nomes de pacotes e capacidades precisam ser verificados no upstream pinado

O adapter central faz resolução de imports por pacote e roteia para runners específicos. Para cada upstream, é necessário validar no SHA pinado: API realmente usada, versão de Python/Node, extras opcionais, requisitos de modelo/credenciais, limites de execução e formato do resultado. Um import disponível não basta para provar operação correta.

## Próxima sequência finita

1. Executar os validadores de registry, ownership e resident contract; compilar runners Python e rodar testes Go do executor.
2. Corrigir apenas defeitos demonstrados nesses gates.
3. Normalizar a semântica dos estados, sem declarar execução live.
4. Auditar raízes históricas duplicadas e o caso `nemo-agent-toolkit`, preservando arquivos até classificar sua função.
5. Criar matriz de compatibilidade por upstream: licença, SHA, runtime, dependências, APIs chamadas, limites, segredos, timeout/cancelamento, operação Mesh e teste.
6. Testar providers por grupos pequenos, começando pelos que têm dependências menores e maior aderência ao owner; desativados por padrão.
7. Só depois validar E2E por núcleo e a integração federada N07 → núcleo proprietário → resultado correlacionado.

## Critério de conclusão

Um upstream só recebe estado **adapter-live** após uma execução real controlada que comprove: SHA de origem conhecido, owner correto, operação permitida, limites aplicados, falha fechada, resposta validada, HMAC/correlação preservados e evidência registrada. CI verde sozinho não satisfaz esse critério.

**Limite desta auditoria:** nenhum upstream foi ativado e nenhuma dependência pública foi atualizada. Este documento registra achados e sequência de correção; não altera `main`.


## Aditivo crítico — frente pública já existente (branch forensic)

A inspeção da branch existente `forensic/public-capability-integration-2026-10-06` revelou que o inventário não termina nos 16 upstreams da `main`. Essa frente declara **32 fontes externas**: 22 fontes primárias/complementares e 10 expansões projetadas. Não abri uma nova PR para isso; esta auditoria é separada e preserva a frente existente.

### Dezesseis fontes adicionais presentes na frente existente

| Fonte adicional | Repositório | Papel declarado | Estado correto nesta auditoria |
|---|---|---|---|
| Autogenesis | [DVampire/Autogenesis](https://github.com/DVampire/Autogenesis) | meta-agente/otimização/avaliação | PROJECTED |
| OctOS | [lispking/octos](https://github.com/lispking/octos) | fan-out paralelo, swarm e sandbox | PROJECTED |
| Hora Graph Core | [Vivien83/hora-graph-core](https://github.com/Vivien83/hora-graph-core) | grafo de conhecimento e memória | PROJECTED |
| Mycelium | [mycelium-io/mycelium](https://github.com/mycelium-io/mycelium) | workspace e coordenação compartilhada | PROJECTED |
| Prime Agent | [PrimeIntellect-ai/prime-agent](https://github.com/PrimeIntellect-ai/prime-agent) | subagentes recursivos/RLM | PROJECTED |
| CUDA Oxide | [SuperInstance/cuda-oxide](https://github.com/SuperInstance/cuda-oxide) | runtime CUDA/PTX e execução GPU | PROJECTED |
| Bijux DAG Runtime | [bijux/bijux-core](https://github.com/bijux/bijux-core) | execução determinística de DAG | PROJECTED |
| Ouro Loop | [VictorVVedtion/ouro-loop](https://github.com/VictorVVedtion/ouro-loop) | loop de verificação e autocorreção | PROJECTED |
| Recuris | [Gen-Verse/Recuris](https://github.com/Gen-Verse/Recuris) | memória experiencial e localização de falhas | PROJECTED |
| FedML | [FedML-AI/FedML](https://github.com/FedML-AI/FedML) | aprendizagem federada e deployment | PROJECTED |
| Hivemind | [learning-at-home/hivemind](https://github.com/learning-at-home/hivemind) | treinamento P2P e agregação distribuída | PROJECTED |
| Temporal | [temporalio/temporal](https://github.com/temporalio/temporal) | workflows duráveis e retry | PROJECTED |
| Cognitive Workspace | [tao-hpu/cognitive-workspace](https://github.com/tao-hpu/cognitive-workspace) | contexto/workspace cognitivo | PROJECTED |
| RAVANA | [OpenSource-Syndicate/RAVANA](https://github.com/OpenSource-Syndicate/RAVANA) | agência proativa e reflexão | PROJECTED |
| Ray | [ray-project/ray](https://github.com/ray-project/ray) | computação distribuída | PROJECTED |
| NATS Go | [nats-io/nats.go](https://github.com/nats-io/nats.go) | transporte de eventos/pub-sub | PROJECTED |

A tabela é um inventário de função declarada, não aprovação técnica. A compatibilidade, licença por pacote, API no SHA fixado, dependências e execução real de cada fonte adicional ainda precisam ser verificadas.

### Falhas e lacunas observadas na execução da frente existente

- **CI de contrato da frente pública:** a execução `40216778045` de “SOUL External Adapter Contract” terminou SUCCESS; a execução `40216777897` de “SOUL External Ownership + Resident Contract CI” também terminou SUCCESS. Isso valida os checks cobertos, não runtime de todos os providers.
- **CI N07:** a execução `40216777945` terminou FAILURE no passo “Production external federation smoke”. O log expõe entradas com `sourcePresent=false`, `sourceState=MISSING_WORKTREE`, `observedRevision=""` para fontes projetadas como Recuris, FedML, Hivemind, Temporal, NATS Go, Ouro Loop, Cognitive Workspace, RAVANA, Ray e Bijux DAG Runtime. Isso prova que a prova de produção não materializou essas fontes no worktree/container usado. **Não prova que o upstream está quebrado.** A pipeline precisa separar “source metadata/probe sem checkout” de “execução que requer código upstream” e garantir que o caminho de runtime só seja executado quando a fonte estiver materializada no SHA esperado.
- **E2E GRCE/SARA:** a execução `40216777915` terminou FAILURE. A asserção mostrou um estado `IDLE`, sem evidence/capabilities, enquanto a entrada contém `TODO: FIXME`. Portanto, o ciclo GRCE→C1-C7-C1 não pode ser considerado concluído; é um bloqueio real da prova de ciclo, separado do inventário dos upstreams.
- **Cancelamento:** a execução `40216778003` cancelou o job de instalação/teste de Ray; isso é `CANCELLED`, não PASS nem FAIL funcional.
- **Pinning divergente entre branches:** várias fontes da lista original têm SHAs diferentes entre `main` e a frente `forensic/public-capability-integration-2026-10-06` (incluindo LangGraph, CrewAI, Microsoft Agent Framework, OpenHands, Browser Use, Letta Code, Pydantic AI, LlamaIndex e DSPy). Cada branch é internamente pinada, mas a integração final deve escolher um SHA canônico por provider e registrar a decisão com evidência. Não atualizar silenciosamente nem escolher “o mais novo” por inferência.

### Correção recomendada para o próximo ciclo

1. Reconciliar o manifesto de 32 fontes com os 16 adapters primários e os 16 adapters/metadata das fontes adicionais; validar IDs, SHA, URL e owner por máquina.
2. Separar três operações: `describe` (metadados, não exige import do upstream), `probe` (presença e compatibilidade, reporta PIN_MATCH/MISSING_WORKTREE/REVISION_MISMATCH) e `execute` (só após preflight, limites e dependências reais).
3. Corrigir o gate que transforma um `MISSING_WORKTREE` projetado em falha ambígua de produção: ele deve falhar se o caso de teste exige execução real, mas não deve rotular metadata-only como execução comprovada.
4. Corrigir a prova GRCE/SARA independentemente; remover TODO/FIXME apenas se o comportamento e as evidências forem substituídos por implementação testada, preservando histórico/proveniência.
5. Reconciliar os SHAs divergentes entre branches antes de promover qualquer adapter a `adapter-live`.
