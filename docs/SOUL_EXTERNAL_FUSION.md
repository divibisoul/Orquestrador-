# SOUL — fusão executável dos 16 upstreams

N07 permanece como plano de controle; a autoridade funcional continua nos núcleos nativos. Os 16 repositórios externos são mantidos como submodules pinados e atravessam uma única fronteira de adapters.

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

Cada upstream possui fonte pinada, adapter explícito, probe, execução por operação, roteamento pelo N07/Mesh, limites de entrada/saída, timeout e falha fechada. A execução externa fica desligada por padrão e é ativada individualmente por variável SOUL_EXTERNAL_EXECUTE_<PROVIDER>=true.

Adapter implementado não é evidência REAL. CI prova contrato e registro; evidência REAL depende de dependências, modelos, credenciais e ambiente compatível.

Operações comuns:
- external.federation.describe@1.0.0
- external.<provider>.probe@1.0.0
- external.<provider>.execute@1.0.0

Nenhum núcleo nativo é substituído e o histórico Git é preservado.
