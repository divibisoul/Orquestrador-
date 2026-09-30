# SUPER AGI — REGRA DE OURO / NOU-RGO / SOUL
## Relatório Mestre de Engenharia — v2
Data: 2026-09-28
Estado: branch de integração isolada, oito PRs draft

### 1. Escopo consolidado
Este documento consolida o trabalho discutido nesta conversa e sua tradução em infraestrutura executável.

Fonte de verdade da implementação: código real nos oito repositórios, seus commits e CI. Mensagens e propostas são material de entrada para auditoria; não são prova de funcionamento.

### 2. Regra da Regra de Ouro
Lei canônica:
F → N(F) → D(F) → C(D(F)) → I → V → H.

- F: erro, falha, lacuna, incompletude, inconsistência ou inadequação observada.
- N(F): natureza determinada por evidência.
- D(F): dual funcional que impede/compensa/transmuta a falha.
- C(D): capacidade implementável.
- I: incorporação.
- V: teste, validação e reauditoria.
- H: preservação de histórico, proveniência e evidência.

Regra de segurança epistemológica:
hipótese ≠ requisito ≠ implementação ≠ teste ≠ validação.
Nenhuma informação é promovida por origem.

Quando o material não contém informação suficiente para determinar o dual, o sistema deve manter DUAL=UNRESOLVED; não preencher a lacuna por imaginação.

### 3. O que foi aprendido nas auditorias anteriores
A etapa forense congelada do NOU-RGO encontrou 24 falhas e transformou as falhas em salvaguardas concretas: integridade de conteúdo, proveniência, evidência material, estados de fechamento, rollback verificado, ciclo de fila, escopo de rede, versões semânticas, consistência de medição, execução real da transmutação, profundidade de auditoria, autorização, evidência de pesquisa e outras.

Essa etapa permanece congelada e não foi reescrita por esta integração.

### 4. Auditoria da proposta BugShield
BugShield foi delimitado como detecção/identificação:
F → N(F) + Evidence.

A auditoria identificou:
- risco de inventar dual no scanner;
- critérios de actionability e utility ainda indefinidos;
- stopping rule sem contrato;
- confusão entre ontologia, epistemologia e modelo;
- PromotionGate sem política independente;
- confusão entre capacidade DERIVED e PROMOTED;
- risco de permanecer no plano teórico sem execução.

Correção estrutural aplicada:
BugShield pode entregar finding sem correction_boundary. O envelope RGO preserva o finding e mantém o dual como UNRESOLVED até uma camada responsável possuir evidência suficiente.

O payload BugShield original é preservado em extensions.bugshield, evitando perda de dados na adaptação.

### 5. Contrato de integração
O contrato RGO v1.0.0 foi implantado em N07 e SARA e adaptadores equivalentes foram implantados em N01–N06.

Estados de capacidade:
DERIVED → IMPLEMENTED → TESTED → VALIDATED → PROMOTABLE → PROMOTED.

Não foram inventados limiares numéricos para UtilityPolicy, StoppingPolicy ou PromotionGate.

### 6. Arquitetura preservada
N01–N06 continuam independentes e usam Soul Mesh 1.1.0 como transporte inter-núcleos.
N07 continua sendo a autoridade de orquestração/controle de fluxo.
SARA continua sendo a fronteira de governança, provenance, regeneração e reauditoria.

HortaCore não foi duplicado:
N01 escreve no HortaCore já existente.
O HortaCore de N07 permanece uma frente histórica separada quando não está em main.

Vagus não foi duplicado:
N01 usa a fachada nervoVago existente sobre o EventBus real.
SARA RGO usa a instância compartilhada de VagusNerveBus existente no bootstrap.
N07 envia para SARA por SARAProxy.

Soul Mesh não foi substituído por Vagus:
Mesh = transporte inter-núcleos.
Vagus = eventos/instrumentação.

### 7. Implantação realizada
N01: RGOFinding + HortaCore/nervoVago + Mesh + testes.
N02: RGOFinding + Mesh + testes.
N03: RGOFinding + Mesh + testes.
N04: RGOFinding + Mesh + testes.
N05: RGOFinding + Mesh + testes.
N06: RGOFinding + Mesh + testes.
N07: envelope RGO, ingestão, adaptador BugShield, SARAProxy RGO, testes e CI.
SARA: RGOEnvelope, RGOEngine, adaptador BugShield, endpoints /v1/rgo/ingest e /v1/rgo/state, bootstrap, testes e CI.

### 8. Proteção do sistema em construção
Todas as alterações desta etapa estão em:
rgo-integration-2026-09-28

Nenhuma alteração foi feita diretamente em main.
Nenhum PR histórico foi reescrito.
Nenhum force-push, reset destrutivo, remoção de feature ou substituição silenciosa foi realizado.
Todos os oito PRs permanecem draft até revisão.

### 9. Evidência de CI — snapshot desta versão
N01 PR #57 — SUCCESS
N02 PR #25 — SUCCESS
N03 PR #23 — SUCCESS
N04 PR #27 — SUCCESS
N05 PR #29 — SUCCESS
N06 PR #22 — SUCCESS
N07 PR #50 — SUCCESS
SARA PR #20 — SUCCESS

Esses resultados comprovam a execução dos workflows RGO correspondentes nos commits registrados. Eles não provam, sozinhos, a integração live de serviços externos que dependem de credenciais, servidores ou infraestrutura não presente no workflow.

### 10. Achado de reauditoria durante a própria implantação
Dois problemas reais foram encontrados e corrigidos pelo próprio ciclo RGO:
1. os runners Node não resolviam corretamente os imports TypeScript dos testes; o dual foi uma execução de teste module-aware e o workflow passou a usar tsx;
2. N01 possuía package-lock.json fora de sincronismo com package.json; o workflow RGO não recebeu uma exigência inventada sobre esse lockfile e passou a verificar a capacidade com instalação compatível, mantendo o problema histórico explicitado.

No N07, um erro de tag JSON introduzido durante a implementação foi detectado pelo CI, corrigido e reexecutado com SUCCESS.

Esse conjunto é a prova prática de autoaplicação: o mecanismo encontrou falhas introduzidas pelo próprio trabalho e as tratou sem ocultá-las.

### 11. Débitos que permanecem visíveis
N06 possui erros TypeScript pré-existentes fora do escopo do adaptador RGO; a verificação do PR RGO é deliberadamente focada no contrato alterado para não mascarar nem reescrever dívida antiga.

Dependências externas que ainda estão fora de main permanecem não-promovidas:
- N07 HortaCore histórico;
- N07 SuperAGI bridge;
- SARA Vagus HTTP;
- N02 VagusBus interno;
- demais frentes Clareira/Mesh/Octacore/parâmetros neurais.

Elas devem ser fusionadas posteriormente por análise de conflito e compatibilidade, nunca por substituição.

### 12. Próxima camada técnica
A próxima integração necessária é ligar o BugShield real à entrada RGO de N07/SARA usando o adaptador já criado, preservando seu schema público v1.2.x.

Depois:
BugShield → RGO Finding → N07/SARA → M01 → M03 → testes/validação/reauditoria → M10.

Os M01–M10 maduros do NOU-RGO não devem ser clonados nos oito repos; devem ser conectados por contrato/adaptador e incorporados como autoridade funcional quando sua implementação real estiver disponível no núcleo correspondente.

### 13. Critério de fechamento
Uma integração só recebe status VALIDATED quando existe implementação, execução, teste, validação, reauditoria e evidência preservada.
“READY” nunca significa “sem erros futuros”.
