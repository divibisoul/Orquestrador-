# SOUL — fusão executável dos 16 upstreams + 6 fontes complementares

O conjunto histórico de 16 upstreams do SOUL-25 permanece preservado. O SOUL-29 acrescenta 6 fontes complementares verificadas, totalizando 22 fontes externas estruturais. As 4 referências solicitadas que retornaram 404 permanecem registradas separadamente como BLOCKED e não são substituídas.

N07 permanece como plano de controle; a autoridade funcional continua nos núcleos nativos. Os 16 repositórios primários e as 6 fontes complementares são mantidos como gitlinks/submodules pinados e atravessam uma única fronteira de adapters.

| Upstream | Dono | Função reforçada |
|---|---|---|
| Superpowers | N07 | engenharia, planejamento, TDD, revisão |
| SuperAGI | N07 | runtime de agentes, ferramentas, memória, multimodalidade |
| LangGraph | N07/N01 | grafos de estado e workflows duráveis |
| CrewAI | N07 | equipes multiagente |
| Microsoft Agent Framework | N07 | agentes, workflows, MCP/A2A |
| OpenHands | N06 | engenharia de software e execução |
| MetaGPT | N07 | processo de software e papéis |
| AgentScope | N03/N07 | agentes, equipes, voz e A2A |
| Letta Code | N01 | memória persistente e identidade |
| Browser Use | N04 | automação de navegador |
| SmolAgents | N06 | agentes de código e ferramentas |
| Pydantic AI | N01/JEV | contratos tipados e subagentes |
| LlamaIndex | N05 | RAG e recuperação |
| DSPy | N06 | raciocínio e otimização |
| Whisper | N03 | fala -> texto |
| Kokoro | N03 | texto -> fala |

Cada upstream possui fonte pinada, adapter explícito, probe, execução por operação, roteamento pelo N07/Mesh, limites de entrada/saída, timeout e falha fechada. Os 16 sources são materializados no checkout de runtime e, na imagem de produção, carregados junto dos adapters e acompanhados por uma atestação de build dos 16 SHAs. A execução externa fica desligada por padrão e é ativada individualmente por variável SOUL_EXTERNAL_EXECUTE_<PROVIDER>=true.

Adapter implementado não é evidência REAL. CI prova contrato e registro; evidência REAL depende de dependências, modelos, credenciais e ambiente compatível.

Operações comuns:
- external.federation.describe@1.0.0
- external.<provider>.probe@1.0.0
- external.<provider>.execute@1.0.0

Nenhum núcleo nativo é substituído e o histórico Git é preservado.


## Ativação por ambiente

- AgentScope: \`SOUL_EXTERNAL_EXECUTE_AGENTSCOPE=true\` + \`OPENAI_API_KEY\`.
- Browser Use: \`SOUL_EXTERNAL_EXECUTE_BROWSER_USE=true\` + \`BROWSER_USE_API_KEY\` ou \`OPENAI_API_KEY\`.
- SmolAgents: \`SOUL_EXTERNAL_EXECUTE_SMOLAGENTS=true\` + \`OPENAI_API_KEY\`.
- Microsoft Agent Framework: \`SOUL_EXTERNAL_EXECUTE_MICROSOFT_AGENT_FRAMEWORK=true\` + \`OPENAI_API_KEY\`.
- SuperAGI: \`SOUL_EXTERNAL_EXECUTE_SUPERAGI=true\` + \`SOUL_EXTERNAL_SUPERAGI_URL\`, \`SOUL_EXTERNAL_SUPERAGI_AGENT_ID\`, \`SOUL_EXTERNAL_SUPERAGI_API_KEY\`.
- OpenHands: \`SOUL_EXTERNAL_EXECUTE_OPENHANDS=true\` + \`SOUL_EXTERNAL_OPENHANDS_URL\`, \`SOUL_EXTERNAL_OPENHANDS_CONVERSATION_ID\`, \`SOUL_EXTERNAL_OPENHANDS_SESSION_KEY\`.
- Letta Code: \`SOUL_EXTERNAL_EXECUTE_LETTA_CODE=true\` + binário \`letta\` e a autenticação/configuração do próprio Letta.

## Fontes complementares do SOUL-29

| Fonte | Estado | Papel |
|---|---|---|
| Autogenesis | PROJECTED | auto-evolução complementar |
| octos | PROJECTED | reforço Octacore / swarm |
| hora-graph-core | PROJECTED | reforço HortaCore / grafo |
| mycelium | PROJECTED | reforço Clareira / workspace |
| prime-agent | PROJECTED | reforço Neocórtex/Clareira / RLM |
| cuda-oxide | PROJECTED | reforço SuperGPU / runtime CUDA |

As fontes são materializadas somente no primeiro nível da federação canônica. Recursão cega em submodules de upstreams é proibida: dependências aninhadas pertencem ao repositório upstream e sua ausência de metadata é preservada como evidência BLOCKED/PROJECTED, não como sucesso fabricado.

### Evidência específica — Autogenesis/HLE

No commit pinado do Autogenesis existe o gitlink `datasets/hle` na revisão `5a81a4c7271a2a2a312b9a690f0c2fde837e4c29`, mas o mesmo commit upstream não fornece `.gitmodules`. A fonte correspondente `cais/hle` existe no Hugging Face, porém é gated; portanto a dependência é registrada, não promovida a execução REAL. A reparação de URL usada na CI é apenas local ao checkout e não altera o upstream.