# SOUL — Auditoria Forense de Recuperação dos Upstreams
## 2026-10-06

### Achado principal

O N07 não perdeu os 16 upstreams. O main atual contém 16 gitlinks com modo 160000 em integrations/external/*, cada um apontando para o SHA pinado no registro de capabilities.

O problema operacional identificado é outro: o bootstrap local não inicializava os submodules externos, embora os adapters apontassem para integrations/external/*. Portanto o conteúdo podia estar registrado no Git e ainda assim não aparecer no working tree.

### Cadeia de preservação

upstream público -> gitlink pinado -> submodule materializado -> SHA verificado -> adapter explícito -> Soul Mesh/N07 -> proprietário nativo

Nenhuma autoridade nativa é substituída. Nenhum segundo Mesh é criado. Nenhum histórico é apagado.

### 16 upstreams

Os 16 fornecedores são: Superpowers, SuperAGI, LangGraph, CrewAI, Microsoft Agent Framework, OpenHands, MetaGPT, AgentScope, Letta Code, Browser Use, SmolAgents, Pydantic AI, LlamaIndex, DSPy, Whisper e Kokoro.

Todos os 16 SHAs atualmente pinados foram confirmados como commits existentes nos respectivos repositórios públicos.

### Drift de upstream

Foi observada divergência entre o SHA pinado e o HEAD público atual em 9 de 16 upstreams: LangGraph, CrewAI, Microsoft Agent Framework, OpenHands, Letta Code, Browser Use, Pydantic AI, LlamaIndex e DSPy.

Esses 9 não são atualizados automaticamente nesta recuperação. Cada promoção será tratada como uma alteração versionada e compatibilidade-gatada, porque atualizar um fornecedor sem testar seu adapter pode introduzir regressão.

### Frentes paralelas preservadas

PRs #123, #124, #125, #126, #127 e #128 continuam independentes e não são alteradas por esta frente. A PR #112 também permanece independente: ela propõe uma segunda onda com 9 upstreams adicionais (ECC, SwarmClaw, Mem0, Letta, LangFuse, vLLM, SGLang, Ray e Megatron-LM).

Esta recuperação nasce do main observado e altera somente o bootstrap/validação do conjunto SOUL-25 já integrado.

### Estado de engenharia

REAL: os 16 upstreams existem como repositórios públicos, os 16 pins são válidos e os 16 gitlinks existem no main.
PROJECTED: runtime de cada provider depende das dependências, credenciais, modelos e ambiente próprios.
BLOCKED: provider ausente, submodule não inicializado ou SHA divergente deve produzir falha explícita.

