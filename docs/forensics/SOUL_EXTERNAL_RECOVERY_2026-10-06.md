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

Os 9 foram promovidos nesta frente para os HEADs públicos observados, sempre como SHA exato e preservando o SHA anterior no histórico e no manifesto. A promoção do pin não é tratada como prova de compatibilidade de runtime; ela permanece subordinada aos contratos, testes e E2E correspondentes.

### Frentes paralelas preservadas

PRs #123, #124, #125, #126, #127 e #128 continuam independentes e não são alteradas por esta frente. A PR #112 também permanece independente: ela propõe uma segunda onda com 9 upstreams adicionais (ECC, SwarmClaw, Mem0, Letta, LangFuse, vLLM, SGLang, Ray e Megatron-LM).

Esta recuperação nasce do main observado e altera somente o bootstrap/validação do conjunto SOUL-25 já integrado.

### Estado de engenharia

REAL: os 16 upstreams existem como repositórios públicos, os 16 pins são válidos e os 16 gitlinks existem no main.
PROJECTED: runtime de cada provider depende das dependências, credenciais, modelos e ambiente próprios.
BLOCKED: provider ausente, submodule não inicializado ou SHA divergente deve produzir falha explícita.



## Pós-recuperação — hardening pela Regra de Ouro

A auditoria posterior encontrou quatro lacunas que ainda impediam chamar a integração de residente no sistema operacional do N07:

1. **Registro de runtime:** os adapters externos existiam, mas não eram registrados pelo `cmd/nexus/main.go`. A correção tornou a federação externa parte efetiva do conjunto de operações do N07.
2. **Integridade de execução:** o runtime confiava no catálogo sem verificar de forma independente `gitlink SHA == registry revision == submodule HEAD`. A correção adicionou essa verificação, bloqueio para worktree/index sujos e falha fechada em divergência.
3. **Fronteira de produção:** a imagem do N07 não carregava os 16 submodules nem o Python adapter. A imagem agora materializa as fontes pinadas, scripts de adapter/runners, Python 3 e uma atestação de build dos 16 SHAs.
4. **Fronteira MultiAgent:** a fachada possuía uma regra histórica de ownership para MetaGPT incompatível com o registro canônico. A correção faz a seleção pelo registry único e reutiliza a mesma fronteira de adapter.

### Estado de evidência após o hardening

- **REAL:** os 16 upstreams estão catalogados como gitlinks exatos; os SHAs promovidos/retidos estão registrados; o N07 registra a federação externa; o caminho de runtime verifica proveniência antes de executar.
- **PROJECTED:** execução real dos providers depende de suas bibliotecas Python/Node, modelos, credenciais, serviços externos e ambiente.
- **BLOCKED:** provider não materializado, revisão divergente, worktree alterado, operação não registrada ou infraestrutura/credencial ausente deve impedir a execução.
- **UNMEASURABLE:** não é inferida a partir de catálogo, pin ou CI qualquer capacidade que não tenha evidência operacional.

A smoke test da imagem foi desenhada para provar residência estrutural da federação e materialização dos 16 sources; ela não falsifica execução autenticada dos 16 providers.
