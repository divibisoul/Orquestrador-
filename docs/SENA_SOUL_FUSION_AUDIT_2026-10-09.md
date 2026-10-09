# SENA + SOUL — auditoria de fusão e fundação aditiva

**Data da inspeção:** 2026-10-09  
**Escopo:** repositórios públicos da organização `divibisoul`, manifests de capacidades, fronteiras de runtime e pull requests relevantes observados.  
**Estado deste artefato:** `SPEC_ONLY`. Esta documentação e os contratos não afirmam que um runtime SENA existe ou foi executado.

## Decisão executiva

SENA deve ser uma **capacidade adaptativa opcional sob coordenação do N07**, não um novo núcleo, outro orquestrador ou outro Mesh. A primeira entrega é uma fundação contratual isolada, para permitir a integração em etapas enquanto as frentes já abertas continuam trabalhando.

O código Python fornecido é uma proposta, não um runtime verificado. A pesquisa de código no `main` não localizou uma implementação SENA própria; os resultados com “SENA” eram correspondências incidentais, como `agentarsenal`. O N07 já contém a rede neural nativa em `neural/network.go`, a montagem da pipeline de aprendizado em `cmd/nexus/main.go`, um gateway federado em `mesh/federated_gateway.go` e um padrão de sidecar com proxy, timeout, token opcional e operações tipadas em `agentarsenal/proxy.go` / `orchestrator/agent_arsenal.go`.

**Diretriz:** SENA primeiro observa e avalia; não substitui decisões, não escreve na memória persistente e não treina online. Só pode avançar depois de contratos, adapter real, proveniência, admissão de recursos, gates de ética, testes e evidência de execução autenticada.

## 1. Mapa real de repositórios SOUL

A listagem do GitHub da organização retornou nove repositórios centrais públicos:

| Repositório | Autoridade funcional a preservar |
|---|---|
| [N01 — aeternum-core-29](https://github.com/divibisoul/aeternum-core-29) | coordenação/estado/identidade/memória, cliente Mesh e fronteiras HortaCore |
| [N02 — Eternium-](https://github.com/divibisoul/Eternium-) | ecossistema multiagente e execução colaborativa |
| [N03 — nexus-aeternum-fusion](https://github.com/divibisoul/nexus-aeternum-fusion) | áudio, fala, voz e percepção multimodal |
| [N04 — nextjs-ai-chatbots](https://github.com/divibisoul/nextjs-ai-chatbots) | interface, ferramentas, documentos e artefatos |
| [N05 — nextjs-ai-chatbot](https://github.com/divibisoul/nextjs-ai-chatbot) | conversa, inferência, conhecimento e recuperação |
| [N06 — nextjs-ai-chatbot-2000](https://github.com/divibisoul/nextjs-ai-chatbot-2000) | cognição, planejamento, raciocínio e agentes de execução |
| [N07 — Orquestrador-](https://github.com/divibisoul/Orquestrador-) | plano de controle, federação, Mesh, SuperGPU e coordenação |
| [SARA](https://github.com/divibisoul/SARA) | ARA/ETR/ITR, ERU/RGO/MMD, regeneração, governança e proveniência |
| [JEV — jev-api](https://github.com/divibisoul/jev-api) | decisões tipadas, triagem e guardrails |

A tabela não declara que toda capacidade esteja comprovada em runtime. Ela registra fronteiras de autoridade que SENA deve respeitar.

## 2. Fontes públicas: reutilizar por afinidade, sem copiar nem duplicar

O manifest atual `integrations/external-capabilities.json` contém 16 fontes externas registradas e revisões pinadas: **Superpowers, SuperAGI, LangGraph, CrewAI, Microsoft Agent Framework, OpenHands, MetaGPT, AgentScope, Letta Code, Browser Use, smolagents, Pydantic AI, LlamaIndex, DSPy, Whisper e Kokoro**. A existência de um pin/catálogo é evidência estrutural; não prova instalação, adapter configurado ou execução autenticada.

O crosswalk em `integrations/sena/public-source-crosswalk.json` mapeia essas fontes às autoridades existentes. O princípio é selecionar funções por afinidade:

- **Agentes e workflows:** LangGraph, CrewAI, Microsoft Agent Framework, SuperAGI e MetaGPT, atrás dos adapters N07/N02/N06 existentes.
- **Engenharia e validação:** Superpowers como metodologia de desenvolvimento; OpenHands e smolagents somente sob execução delimitada.
- **Memória e recuperação:** Letta Code para um adapter de agente com estado, LlamaIndex por N05 para RAG. Nenhum deles toma a autoridade de estado de N01 nem a proveniência/regeneração de SARA.
- **Linguagem e otimização:** Pydantic AI para contratos tipados, DSPy para avaliação/otimização offline e não para alteração online autônoma de políticas.
- **Percepção:** Whisper, Kokoro e AgentScope por N03; Browser Use e processamento de documentos por N04; recuperação por N05.

### Fontes adicionais nas frentes abertas

Os seguintes grupos estão em branches/PRs observados, não devem ser apresentados como incorporados ou ativos em produção somente por constarem desta auditoria:

| Frente | Grupo público documentado | Observação de integração |
|---|---|---|
| [N07 #112](https://github.com/divibisoul/Orquestrador-/pull/112) | ECC, SwarmClaw, Mem0, Letta, Langfuse, vLLM, SGLang, Ray, Megatron-LM | stack de agentes/memória/avaliação/GPU; exige pins e adapter real |
| [N07 #129](https://github.com/divibisoul/Orquestrador-/pull/129) | recuperação/materialização de SOUL-25, incluindo fontes como AutoGenesis, CUDA-Oxide, HORA Graph Core, Mycelium, Octos e Prime Agent | verificar os gitlinks, worktrees e gates sem assumir materialização operacional |
| [N07 #144](https://github.com/divibisoul/Orquestrador-/pull/144) | onda upstream GRF/GRCE, incluindo CUDA-Oxide, HORA Graph Core, Mycelium, Octos, OpenMind Conductor e Prime Agent | usar o ciclo/invariantes GRF/GRCE como gates, sem criar outro ciclo de autoridade |
| [N07 #146](https://github.com/divibisoul/Orquestrador-/pull/146) | Bijux DAG Runtime, Ouro Loop e Recuris | extensão SOUL-28 baseada em #144; preservar a extensão como ramo próprio |
| [N07 #152](https://github.com/divibisoul/Orquestrador-/pull/152) | Bijux DAG Runtime, Cognitive Workspace, FedML, Hivemind, NATS Go, Ravana, Ray, Recuris e Temporal | branch baseado na recuperação #129; não importar manifests isoladamente da sua cadeia |

Isso cobre as fontes upstream que já aparecem no inventário principal e nas frentes públicas registradas. **Não recomendo aumentar a contagem de dependências sem uma lacuna de capacidade demonstrada.** O ganho do SENA deve vir primeiro da composição correta do que já está catalogado.

## 3. Pontes entre SENA e as frentes existentes

| Capacidade pretendida no SENA | Ponte correta no SOUL | Limite não negociável |
|---|---|---|
| NLP, idioma e políticas de ação | N05 para inferência/recuperação; N02/N06 para execução cognitiva e agentes | não manter uma segunda autoridade de geração/planejamento |
| Multimodal | N03 (áudio/voz), N04 (ferramentas/documentos), N01 (contexto de dispositivo autorizado) | nenhum tensor de imagem/sensor aleatório como dado real |
| Rede neural/aprendizado | rede neural Go e pipeline `learning` de N07; revisar a frente [#130](https://github.com/divibisoul/Orquestrador-/pull/130) | escolher contratos e uma autoridade de treinamento; evitar segundo otimizador concorrente |
| Memória associativa | interfaces existentes de estado/memória N01, retrieval N05 e proveniência/trace SARA | leitura inicialmente; sem segunda memória persistente ou escrita direta no banco |
| Restrições éticas | SARA ARA → ETR → ITR, evidência ERU/RGO/MMD e guardrails tipados JEV | score neural nunca autoriza uma ação por conta própria |
| Recursos ociosos | admissão e escalonamento existentes de N07/SuperGPU/Octacore/Prefrontal | não usar apenas percentuais fixos de CPU/GPU; medir RAM, VRAM, fila, deadline e SLO do host |
| Mensageria e resposta | contrato único Soul Mesh 1.1.0, identidade/correlação e gateway N07 | nenhum segundo Mesh; VagusBus é transporte de sinais, não Mesh |
| Estado/evidência durável | HortaCore/N01 e proveniência SARA; acompanhar a frente [N07 #125](https://github.com/divibisoul/Orquestrador-/pull/125) | não afirmar persistência E2E sem hashes e confirmação do peer real |

### Encadeamento a preservar

1. O consumidor entra pelo contrato de Mesh existente e apresenta `trace_id`, `correlation_id`, deadline e payload validado.
2. N07 resolve/admite a capacidade e delega por adapter com timeout, limite de payload e revisão da fonte registrada.
3. SENA devolve uma **sugestão tipada**, com hash de entrada/saída, versão do modelo/fonte, custo e estado epistêmico.
4. Se a saída puder afetar decisão, SARA/JEV e os invariantes aplicáveis precisam aprová-la. Falha, ausência de configuração e evidência insuficiente permanecem explícitas.
5. Apenas os caminhos atuais podem persistir estado ou emitir sinais: HortaCore para estado, VagusBus para sinal, Mesh para transporte/federação.

## 4. Defeitos bloqueadores no código SENA proposto

Além dos problemas resumidos na revisão inicial, a leitura estática do snippet revela:

- `star_mask` é calculada, mas não usada no `MultiheadAttention`; a topologia declarada não controla as conexões reais.
- `NucleoMultimodal.text_proj` aceita 768 dimensões, mas recebe `nlp_emb` projetado para 512 por padrão.
- `q_values` tem dimensão de ações (8), não `d_model` (512); empilhá-lo com embeddings falha.
- A política tem quatro ações e o DQN oito. `policy_loss` multiplica log-probabilidades por `batch["action"]` sem um contrato correto de índices/one-hot; a semântica de `next_q` e dos alvos também não está fechada.
- O `learn()` chama `forward()`, que grava memória como efeito colateral; um treinamento/retry pode alterar estado de memória em momento impróprio.
- `MetaLearner` copia gradientes e os aplica de novo depois do otimizador; não realiza, por si só, MAML/Reptile. O estado de momentum desse meta-aprendizado também não está no `state_dict`.
- `NucleoEtico.lambdas` são parâmetros irrestritos: podem ficar negativos. Um score aprendido não constitui enforcement de política.
- `update_fisher()` não faz parte de um ciclo EWC completo e atualiza buffers via `.data`; a implementação precisa ser substituída e testada antes de reivindicar consolidação EWC.
- A memória chamada “holográfica” é, no snippet, uma tabela de slots com atenção por similaridade; não comprova as propriedades descritas de memória holográfica nem consolidação histórica.
- `_prepare()` cria imagens e sensores com `torch.randn`: isso é dado artificial, não observação do host.
- `observe()` roda depois de o host terminar e de forma síncrona; não garante paralelismo nem ausência de impacto de latência. `tensor.numpy()` também exige cuidados com dispositivo e gradientes.
- O gerenciador de recursos não mede RAM/VRAM, pressão de memória, fila, deadline nem admissão do escalonador central. Os limites 60%/70% não provam “energia ociosa”.
- `torch.load` pode carregar conteúdo baseado em pickle. Checkpoints não confiáveis exigem formato restrito e validação; não usar pickle arbitrário como fronteira de segurança.
- `comparison_log` e `historico` podem crescer sem limite e reter prompts/respostas sensíveis. São necessários redaction, minimização, retenção e armazenamento de evidência controlado.
- O modo produção retorna o resultado do SENA em vez de garantir que o output original do host permaneça autoritativo. O kill switch e o fallback no snippet não tornam essa mudança segura.
- `_compativel()` sempre retorna `True` e `_combinar()` retorna o host sem combinar; essas rotinas não implementam a arbitragem alegada.
- `Ray` é importado, mas não há implementação concreta de pool de atores/agentes que prove a arquitetura de cinco agentes.

## 5. Estado de CI das frentes relacionadas observado nesta auditoria

O snapshot não está verde para promoção geral:

- [N07 #152](https://github.com/divibisoul/Orquestrador-/pull/152): a verificação pública apresentou falha na instalação das dependências de Recuris; o job de Ray foi cancelado. Também falharam o passo `Validate provider-to-adapter graph` do contrato de adapters e o passo `Validate Soul-29 and GRF invariants` da integridade N07. Referências: [workflow de fontes públicas](https://github.com/divibisoul/Orquestrador-/actions/runs/37551677734), [contrato de adapters](https://github.com/divibisoul/Orquestrador-/actions/runs/37551677793), [CI N07](https://github.com/divibisoul/Orquestrador-/actions/runs/37551677876).
- [N07 #153](https://github.com/divibisoul/Orquestrador-/pull/153): o passo `Test` do CI N07 e o passo de execução E2E real GRCE → SARA falharam na amostra de runs consultada. A falha da etapa foi observada; a causa de raiz não é declarada aqui sem análise dos logs. Referências: [CI N07](https://github.com/divibisoul/Orquestrador-/actions/runs/37919071032), [E2E GRCE/SARA](https://github.com/divibisoul/Orquestrador-/actions/runs/37919071136).

Por isso, esta fundação não altera nem mascara essas falhas, não fecha PRs e não promove nenhum upstream a `REAL`. A meta de “CI verde” precisa ser cumprida em cada branch e na integração final, não presumida.

## 6. Plano de implantação seguro

**Fase A — contrato/fundação (esta entrega).** Adicionar somente os arquivos isolados `integrations/sena/*`, a documentação e um gate de validação próprio. Não editar o registry central, `.gitmodules`, rede neural nativa ou fronteiras de outras frentes.

**Fase B — adapter N07.** Reutilizar o padrão de proxy/sidecar já presente: configuração explícita, timeout, limite de resposta, token, erro tipado e correlação preservada. Desligado por padrão; sem monkey-patching de métodos do host. Health indica saúde do processo, não prova de execução.

**Fase C — modo sombra real.** Consumir somente payload autorizado e real, com schema versionado. Registrar hash, latência, custo, decisão host e sugestão SENA de forma minimizada. A saída do host continua idêntica; nenhuma ação externa nem treinamento online. A operação deve ser cancelável e sujeita a um orçamento de recursos do N07.

**Fase D — avaliação offline.** Antes de RL, fechar action space, estado, recompensa versionada, dados de treino/validação, métricas, avaliação fora de amostra e testes de regressão. Comparar primeiro a capacidade neural nativa atual com o candidato SENA para decidir o que deve ser delegado ou reaproveitado.

**Fase E — assist controlado.** Somente com compatibilidade semântica real, calibração de confiança, evidência de execução e autorização SARA/JEV. A sugestão permanece consultiva; divergências e vetos são preservados como evidência.

**Fase F — promoção.** Requer CI verde, E2E autenticado N07 ↔ SENA ↔ SARA/JEV nos cenários aplicáveis, proveniência de ponta a ponta, limites de memória/recursos, rollback testado e aprovação humana explícita. Produção continua bloqueada até esses gates passarem.

## 7. Critérios de aceite

- Nenhum arquivo existente é removido, reescrito ou renomeado por esta frente.
- Nenhum novo núcleo, Mesh, autoridade de memória ou ciclo regenerativo é criado.
- SENA está desativado por padrão e o host segue como fonte de verdade no modo sombra.
- A entrada e a saída preservam trace/correlation IDs, hashes, revisão de origem e estado `REAL/PROJECTED/BLOCKED/UNMEASURABLE`.
- Dados aleatórios nunca são apresentados como dados reais; erros nunca são convertidos em sucesso sintético.
- Fontes externas permanecem pinadas e separadas do runtime; adapters não verificados permanecem bloqueados.
- A promoção depende das frentes #129/#152, #144/#146, #130, #153, #43 e #125 sem substituir suas responsabilidades ou alterar seus branches.

## Referências de código inspecionadas

- [Manifesto de capacidades externas](https://github.com/divibisoul/Orquestrador-/blob/main/integrations/external-capabilities.json)
- [Ownership das capacidades externas](https://github.com/divibisoul/Orquestrador-/blob/main/integrations/external-capability-ownership.json)
- [Contrato dos componentes SOUL](https://github.com/divibisoul/Orquestrador-/blob/main/integrations/soul-component-contracts.json)
- [Rede neural nativa Go](https://github.com/divibisoul/Orquestrador-/blob/main/neural/network.go)
- [Inicialização N07 e pipeline de aprendizado](https://github.com/divibisoul/Orquestrador-/blob/main/cmd/nexus/main.go)
- [Gateway federado Mesh](https://github.com/divibisoul/Orquestrador-/blob/main/mesh/federated_gateway.go)
- [Proxy de sidecar existente](https://github.com/divibisoul/Orquestrador-/blob/main/agentarsenal/proxy.go)
- [N07 → operações do sidecar](https://github.com/divibisoul/Orquestrador-/blob/main/orchestrator/agent_arsenal.go)
- [Kernel G0 existente de SARA](https://github.com/divibisoul/SARA/blob/main/src/sara/meta/octacore_kernel.py)
- [Boundary contratual de SARA](https://github.com/divibisoul/SARA/blob/main/integrations/capability-boundary.json)
