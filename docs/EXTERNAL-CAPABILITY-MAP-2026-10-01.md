# AETERNUM/SOUL — Mapa de Capacidades Externas
**Data:** 2026-10-01  
**Estado:** MAPA ADITIVO / INTEGRAÇÃO CONTROLADA  
**Repositório:** divibisoul/Orquestrador-  
**Branch:** `integration/external-capability-map-2026-10-01`

## 1. Propósito

Os 15 repositórios externos são tratados como **fontes de capacidades** para completar o sistema AETERNUM/SOUL/SARA. O objetivo não é adotar um framework inteiro por preferência; é localizar funções reais, comparar com as capacidades já existentes em N01–N07/SARA e incorporar somente onde houver ganho funcional verificável.

A regra operacional permanece:

> NÃO EXCLUIR. NÃO INVALIDAR. NÃO SIMPLIFICAR. NÃO SOTERRAR. NÃO SUBSTITUIR. NÃO SIMULAR. NÃO ALTERAR ARQUITETURA. APENAS AUDITAR. APENAS RECUPERAR. APENAS INTEGRAR. APENAS COMPLETAR. APENAS CORRIGIR. APENAS OTIMIZAR.

### 1.1 Princípio de preservação

1. O repositório externo pode permanecer inteiro na área de aquisição.
2. A integração em produção pode ser **inteira**, **parcial**, ou **por adapter**, conforme arquitetura, dependências e licença.
3. Nunca apagar o original para incorporar um subconjunto.
4. Toda incorporação deve registrar: fonte, commit/versão, caminho, licença, transformação aplicada, proprietário funcional e prova de testes.
5. Nenhuma capacidade é considerada ativa somente porque foi copiada. Estado ativo exige evidência de execução e integração real.
6. Mocks/simulações/sucessos sintéticos não servem como prova.

---

## 2. Classes de capacidade

| ID | Classe | O que procuramos |
|---|---|---|
| C01 | Agentes e execução autônoma | ciclo agente, execução iterativa, subagentes, concorrência |
| C02 | Planejamento e decomposição | goal→plan→act, filas, tarefas, workflows, replanejamento |
| C03 | Ferramentas e descoberta | tool registry, toolkits, MCP, descoberta/ativação dinâmica |
| C04 | Memória e recuperação | memória de longo prazo, FTS/BM25, vetorial, temporal, grafos |
| C05 | Conhecimento semântico | grafos, KG, Graph2Seq/Graph2Tree, semantic parsing |
| C06 | Aprendizagem online/adaptação | incremental learning, drift, monitoramento, atualização contínua |
| C07 | Runtime neural/ML | treinamento, inferência, pipelines, tensores, aceleradores |
| C08 | Sandbox/execução segura | shell/código, isolamento, permissões, artefatos |
| C09 | Engenharia autônoma | skills, hooks, TDD, revisão, planos, worktrees |
| C10 | Observabilidade/eval | telemetria, avaliações, traces, métricas, diagnóstico |
| C11 | Governança/safety | políticas, segurança, controle, compliance, safety research |
| C12 | Pesquisa/meta-conhecimento | papers, roadmaps, frameworks, recursos, taxonomias |
| C13 | Fundamentos algorítmicos | estruturas de dados, algoritmos, complexidade, padrões |
| C14 | Integrações/API/federação | gateways, adapters, providers, serviços externos |
| C15 | Multimodalidade/ambiente | imagem, áudio, navegador, dados externos, percepção |

---

## 3. Agrupamento por semelhança entre os núcleos

### G1 — Agente + Planejamento + Execução
**N02 ↔ N07 ↔ SARA**, com apoio de N05/N06.

Fontes principais: AutoGPT, DeerFlow, BabyAGI, SuperAGI, Rasa.

- N02: raciocínio/inferência, conversação e execução de ferramentas.
- N07: composição, delegação, workflow, federation e execução distribuída.
- SARA: auditoria/regeneração/governança do ciclo, não substituição do executor.
- N05/N06: ferramentas, memória, capabilities e adapters reutilizáveis.

### G2 — Conhecimento + Memória + Semântica
**N03 ↔ N04 ↔ SARA**, com apoio de N02.

Fontes principais: Graph4NLP, AutoGPT/Graphiti, DeerFlow memory/knowledge, AGI-Papers.

- N03: conhecimento estruturado, multimodalidade e memória semântica.
- N04: documentação, contexto, streaming e interfaces de conhecimento.
- SARA: proveniência, pesquisa, assimilação e regeneração.
- N02: consulta/uso dessas capacidades durante inferência.

### G3 — Ferramentas + Skills + Engenharia autônoma
**N05 ↔ N06 ↔ N02**, com fronteira N01.

Fontes principais: DeerFlow tool search/MCP, Superpowers, Everything Claude Code, AutoGPT tools, SuperAGI toolkits.

- N05: artefatos, tools, memória e integração.
- N06: CapabilityEngine/ToolRegistry/adapters/skills.
- N02: descoberta e chamada das ferramentas durante raciocínio.
- N01: contrato, autoridade e Mesh; não receber lógica duplicada de execução.

### G4 — Aprendizagem contínua + Adaptação
**N06 ↔ N07 ↔ SARA**, com apoio de N03.

Fontes principais: River, TensorFlow, partes de AutoGPT/DeerFlow com monitoramento.

- N06: modelos/pipelines de aprendizagem e adaptação.
- N07: aplicação de políticas de adaptação no ciclo federado.
- SARA: auditoria, governança, snapshots e regeneração.
- N03: sinais/conhecimento usados para adaptação.

### G5 — Segurança + Verificação + Governança
**N01 ↔ N07 ↔ SARA**, com participação de todos os owners.

Fontes principais: DeerFlow sandbox/authz, Everything Claude Code security/eval, SuperAGI telemetry, Awesome AGI/ACI/ASI, AGI-Papers.

- N01: contratos, fronteiras e integridade.
- N07: autorização/roteamento/delegação.
- SARA: governança, ética, proveniência e regeneração.
- N02–N06: controles locais do capability owner.

### G6 — Fundamentos de engenharia e algoritmos
**N06 ↔ SARA**.

Fontes principais: javascript-algorithms, build-your-own-x, coding-interview-university.

Estas fontes não devem ser tratadas automaticamente como runtime. Elas fornecem implementações de referência, algoritmos, padrões e material para **Skill Acquisition / Engineering / Meta-learning**.

---

## 4. Matriz fonte → capacidade → owner

| Fonte | Capacidades verificadas | Owner primário | Owners secundários | Forma de integração |
|---|---|---|---|---|
| Graph4NLP | C05, C04, C02 | N03 | N04, SARA | módulos seletivos/adapter; preservar pipeline de referência |
| Rasa | C01, C02, C03 | N02 | N04, N07 | componentes seletivos para diálogo/policy/action |
| River | C06, C07, C10 | N06 | N07, SARA | biblioteca/adapters; extrair estimadores e drift |
| AutoGPT | C01, C02, C03, C04, C08, C10, C14, C15 | N02 | N03, N05, N06, N07 | **não** copiar `autogpt_platform` indiscriminadamente; separar partes por licença |
| DeerFlow | C01, C02, C03, C04, C08, C10, C14 | N07 | N02, N03, N05, N06 | extrair mecanismos específicos (tool search/MCP/memory/sandbox) |
| SuperAGI | C01, C02, C03, C04, C10, C14, C15 | N07 | N02, N05, N06, N03 | toolkit/workflow/memory/telemetry; MIT conforme escopo da origem |
| BabyAGI | C01, C02, C03, C09 | N07 | N02, N06 | estudar e extrair planner/task/function primitives; arquivo original preservado |
| Superpowers | C09, C10 | N06 | N02, N05 | importar skills/metodologia, não a harness como nova autoridade |
| Everything Claude Code | C03, C09, C10, C11 | N06 | N02, N05, SARA | skills/rules/hooks/eval/security como capabilities/knowledge |
| build-your-own-x | C13, C07, C14 | N06 | SARA | fonte de referência/skill acquisition; não runtime direto |
| TensorFlow | C07, C06 | N06 | N07, SARA | somente subcomponentes necessários; não incorporar a stack inteira |
| javascript-algorithms | C13 | N06 | SARA | algorithms/reference skills; seleção por demanda |
| coding-interview-university | C13, C09, C12 | N06 | SARA | currículo/meta-learning, não runtime |
| awesome-agi-aci-asi | C11, C12, C14 | SARA | N07, N06 | corpus de pesquisa/taxonomia/safety |
| AGI-Papers | C12, C11, C04 | SARA | N03, N04, N07 | ingestão estruturada de pesquisa; proveniência obrigatória |

---

## 5. Achados de código relevantes já confirmados

### DeerFlow
- `tool_search.py`: catálogo imutável de ferramentas, busca por schema/nome, promoção de ferramentas diferidas e hash do catálogo; há lógica explícita de fail-closed na montagem de ferramentas.
- `deermem/core/retrieval.py`: FTS5/BM25, ranking com tempo/confiança, categorias e isolamento por usuário/agente.
- `sandbox/sandbox.py`: abstração de sandbox para execução isolada.
- `mcp/tools.py`: superfície dedicada para ferramentas MCP.

### AutoGPT
Foram localizados:
- ferramentas base e sandbox;
- integração Graphiti para memória/knowledge;
- estratégias de prompt/planning;
- blocos/componentes de execução;
- integrações e ferramentas MCP.

**Restrição importante:** o próprio LICENSE atual separa `autogpt_platform/` sob Polyform Shield e o restante sob MIT. A incorporação deve respeitar o limite de licença por caminho.

### River
- `drift/binary/hddm_a.py`: detector de drift online.
- `compose/pipeline.py`: pipelines incrementais e roteamento de parâmetros entre etapas.
- O conjunto é especialmente aderente à camada de adaptação/monitoramento do N06.

### Superpowers
A implementação é organizada como skills composáveis para agentes de engenharia e inclui um fluxo operacional de brainstorming → design → plano → worktree → desenvolvimento por subagentes → revisão/testes.

### SuperAGI
A origem contém agentes concorrentes, toolkits, workflows, memória de agente, múltiplos vector stores, telemetry, token optimization e integrações externas.

---

## 6. Regra para copiar inteiro vs. extrair

### COPY-WHOLE / preservação
Pode ser feito na área de aquisição quando o objetivo é manter a fonte intacta e versionada para arqueologia, comparação e futura recuperação.

### EXTRACT-SELECTIVE / integração de produção
Preferido quando:
- o repositório é muito maior que a capacidade necessária;
- existe runtime próprio já ativo em AETERNUM;
- há dependências conflitantes;
- o mecanismo externo é uma implementação específica de uma capability já existente;
- a licença diferencia partes do repositório.

### ADAPTER
Preferido quando a funcionalidade é valiosa, mas a autoridade de execução já pertence a N01–N07/SARA.

**Nunca criar um segundo runtime paralelo apenas porque uma fonte externa possui uma implementação semelhante.**

---

## 7. Ordem técnica da aquisição

1. **Preservar** as 15 fontes completas na área de staging.
2. Fixar **commit SHA** de cada fonte.
3. Registrar licença e notices.
4. Extrair inventário de capabilities por arquivo/módulo.
5. Comparar com os registries e owners reais N01–N07/SARA.
6. Para cada capability, decidir: `WHOLE`, `SELECTIVE`, `ADAPTER` ou `REFERENCE`.
7. Integrar no **owner nativo**, mantendo o executor original e o Mesh canônico.
8. Rodar testes unitários/integrados e verificar artefatos.
9. Somente depois realizar prova online real.
10. Manter origem, diff, SHA, evidências e falhas para permitir reversão/reprodução.

---

## 8. Primeiras integrações concretas a investigar

### Frente A — Tool Discovery / MCP
**Fonte:** DeerFlow  
**Destino:** N06 → N02, com rota por N07.

Objetivo: comparar `DeferredToolCatalog`, promoção por thread e hash do catálogo com ToolRegistry/CapabilityEngine já existentes. Não criar um segundo registry; reaproveitar o mecanismo onde ele cobrir uma lacuna real.

### Frente B — Memória de recuperação
**Fonte:** DeerMem  
**Destino:** N03/N05  
**SARA:** proveniência e auditoria.

Objetivo: comparar BM25 + time-decay + confidence + scope isolation com a memória vetorial/temporal existente. Extrair somente o mecanismo ausente, preservando a memória nativa.

### Frente C — Adaptação online
**Fonte:** River  
**Destino:** N06  
**N07/SARA:** política, monitoramento e governança.

Objetivo: incorporar detectores de drift e pipelines online como capability adapters, conectando-os ao mecanismo de adaptação existente.

### Frente D — Planner/agent loop
**Fontes:** BabyAGI + AutoGPT + DeerFlow + SuperAGI  
**Destino:** N07/N02.

Objetivo: comparar decomposição de tarefas, subagentes, filas, workflows e execução com o Goal→Plan→Act e o orquestrador existentes. Reaproveitar primitives; não criar um novo orchestrator paralelo.

### Frente E — Engenharia autônoma
**Fontes:** Superpowers + Everything Claude Code  
**Destino:** N06/N05/N02.

Objetivo: transformar skills, hooks, regras de execução, avaliação e segurança em capabilities reutilizáveis do sistema, sem acoplar o SOUL a um harness externo.

### Frente F — Conhecimento semântico
**Fontes:** Graph4NLP + Graphiti/knowledge de AutoGPT  
**Destino:** N03.

Objetivo: reforçar construção/consulta de grafos e raciocínio sobre relações, comparando com o conhecimento existente.

### Frente G — Pesquisa e safety
**Fontes:** AGI-Papers + Awesome AGI/ACI/ASI  
**Destino:** SARA.

Objetivo: alimentar `QuantumCrawler`, `InnovationRadar`, `AssimilationReviewCommittee` e mecanismos de proveniência com material estruturado; não tratar listas/documentação como código executável.

---

## 9. Estado

**REAL / VERIFICADO:** existência atual das fontes e localização de múltiplos mecanismos relevantes nos repositórios consultados.

**PROJETADO:** o mapeamento source → capability → nucleus acima.

**AINDA NÃO PROVADO:** integração funcional desses mecanismos no runtime AETERNUM/SOUL em produção.

**BLOQUEIO ATUAL:** cada integração precisa passar por compatibilidade de arquitetura, dependências, licença e testes no owner nativo.

**UNMEASURABLE:** qualquer alegação de ganho de inteligência antes de benchmark/prova operacional correspondente.

---

## 10. Regra de sinergia

A unidade de integração não é o repositório. É a **capacidade**.

Uma mesma capacidade pode ter implementações complementares em fontes diferentes. O sistema deve preservar as implementações e selecionar/compor a implementação apropriada no owner nativo, registrando a proveniência.

Exemplo:

`Planner`
= primitives BabyAGI
+ estratégias AutoGPT
+ composição/subagentes DeerFlow
+ workflow/tooling SuperAGI
+ execução/autoridade N07
+ auditoria/regeneração SARA.

Isto é **composição verificável**, não cópia cega nem substituição de arquitetura.
