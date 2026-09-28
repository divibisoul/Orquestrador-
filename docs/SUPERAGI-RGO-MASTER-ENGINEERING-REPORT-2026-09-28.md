# SUPER AGI — REGRA DE OURO / NOU-RGO / SOUL
## Relatório Mestre de Engenharia e Implantação
Data: 2026-09-28
Escopo: integração aditiva nos oito núcleos (N01–N07 + G0/SARA)

## 1. Regra de não-presunção
Todo material recebido é um objeto de auditoria. Mensagem, prompt, código, proposta, analogia ou relatório não ganha autoridade por origem. Cada afirmação mantém seu estado epistemológico até existir evidência suficiente.

Estados principais:
FACT/DATA, EXECUTION, INSPECTION, INFERENCE, HYPOTHESIS, PROPOSED, VALIDATED, REJECTED, INCONCLUSIVE.

## 2. Lei canônica implantada como contrato
Todo erro contém, em sua negação, a estrutura que o impede. A Lei da Regra de Ouro extrai essa estrutura e a incorpora ao sistema como infraestrutura permanente.

Ciclo:
F → N(F) → D(F) → C(D(F)) → I → V → H

F = falha/erro/gap/incompletude/inconsistência.
N = natureza diagnosticada por evidência.
D = dual funcional; não é antônimo textual nem teoria inventada.
C = capacidade mínima que materializa o dual.
I = incorporação.
V = teste/validação/reauditoria.
H = histórico/proveniência preservados.

Quando o dual não puder ser derivado de um correction boundary explícito, o estado deve ser DUAL_UNRESOLVED. A ausência de dados não é preenchida por inferência.

## 3. Regras de conservação
Nenhum código existente é removido. Nenhuma funcionalidade histórica é apagada para instalar uma nova. A integração usa adaptadores e enlaces. Sistemas com autoridade prévia permanecem autoridades.

Soul Mesh 1.1.0 = transporte inter-núcleos.
Vagus/VagusNerveBus = evento/instrumentação/control plane local.
HortaCore = memória/substrato.
SARA = governança, regeneração, provenance e fechamento do ciclo.
N07 = orquestração/controle de fluxo.
Jev/SuperAGI = sistemas externos via adaptadores, sem duplicação interna.

## 4. Descobertas estruturais das auditorias anteriores
A auditoria forense congelada do NOU-RGO encontrou 24 falhas e transformou cada uma em salvaguarda/capacidade, incluindo integridade de conteúdo de eventos/proveniência, gate de evidência material, fechamento explícito, rollback verificado, ciclo de fila, escopo de rede, versões semânticas, consistência de amostras, evidência de adaptação, honestidade da transmutação, profundidade forense, limites de execução, allowlist de autorização, exigência material de evidência, exit code de bloqueio, profundidade semântica, execução real do ciclo de ação, integridade da fila e ciclo de pesquisa verificável.

Essa etapa permanece congelada; sua infraestrutura é referência, não será reescrita por esta branch.

## 5. Aprendizado aplicado ao BugShield
BugShield é camada especializada de detecção:
F → N(F) + Evidence.

Ele não inventa D(F), C(D) nem autoridade de fechamento. Entrega finding estruturado, evidências, cobertura honesta, proveniência, estado epistêmico e limitações para M01/SARA/N07.

A auditoria da proposta BugShield identificou e corrigiu conceitualmente os seguintes riscos:
- critérios indefinidos de erro acionável;
- critérios de utilidade inventados;
- stopping rule inventada;
- confusão ontologia/epistemologia/modelo;
- PromotionGate sem contrato externo;
- confusão entre capacidade derivada, implementada, validada e promovida;
- risco de permanecer em projeto teórico sem execução.
A forma implantada evita limiares inventados e preserva estados UNDEFINED/INCONCLUSIVE/UNRESOLVED.

## 6. Fluxo definitivo de engenharia
Entrada → preservação → classificação epistemológica → auditoria → finding → causa/natureza/impacto → dual funcional → capacidade mínima → implementação → execução → teste → validação → reauditoria → promoção ou rejeição → histórico.

Uma etapa não retroativamente cria evidência de outra.
Hipótese ≠ requisito ≠ implementação ≠ teste ≠ validação.

## 7. Ciclo de capacidade
DERIVED
→ IMPLEMENTED
→ TESTED
→ VALIDATED
→ PROMOTABLE
→ PROMOTED

PromotionGate somente executa uma política/contrato previamente definido e versionado. Ele não define sozinho seus critérios.

UtilityPolicy e StoppingPolicy podem existir como contratos, mas nenhum limiar numérico é inventado pelo runtime.

## 8. Implantação nos oito núcleos
N01: adapter real que usa HortaCore existente, nervoVago/EventBus existente e Soul Mesh.
N02: adapter RGO nativo sobre Soul Mesh.
N03: adapter RGO nativo sobre Soul Mesh.
N04: adapter RGO nativo sobre Soul Mesh.
N05: adapter RGO nativo sobre Soul Mesh.
N06: adapter RGO nativo sobre Soul Mesh.
N07: contrato Go, operação rgo.ingest@1.0.0 e fronteira para SARA por SARAProxy.
G0/SARA: RGOEngine, provenance, VagusNerveBus existente, hash chain, endpoints /v1/rgo/ingest e /v1/rgo/state.

## 9. HortaCore e VagusBus
N01 já possui HortaCore real em memória e uma fachada nervoVago sobre o EventBus real. A integração RGO escreve no HortaCore existente e emite pelo nervoVago.
SARA possui VagusNerveBus in-process. O RGOEngine usa a instância compartilhada criada no bootstrap e não cria um novo protocolo de barramento.
O HortaCore de N07 que está em PR aberto não é importado de branch externa nesta implantação; isso evita depender de código não presente em main.

## 10. Dependências abertas observadas
Recuperação histórica HortaCore em N07: PR #47, ainda separado de main no momento da auditoria.
Integração SuperAGI N07: PR #49, com ponte fail-closed; execução live não é declarada sem endpoint/credencial.
Vagus HTTP SARA: PR #17, separado de main; esta branch RGO adiciona seu próprio endpoint RGO sem reescrever PRs existentes.
N02 VagusBus interno: PR #24, separado; esta branch não substitui nem absorve esse design.
Outras PRs de recuperação, Mesh, Octacore e parâmetros neurais permanecem independentes.

## 11. Segurança de integração
Esta implantação foi criada em branch isolada rgo-integration-2026-09-28 em todos os oito repositórios. Nenhum main recebeu alterações. Não houve force-push, reset destrutivo ou remoção de arquivo. Cada PR é draft e pode ser revisado/reconciliado antes de qualquer merge.

## 12. Evidência de execução observada nesta etapa
N01 RGO CI: SUCCESS.
N02 RGO CI: SUCCESS após corrigir a limitação do runner TypeScript.
N04 RGO CI: SUCCESS.
N05 RGO CI: SUCCESS.
N07 RGO CI: SUCCESS.
SARA RGO CI: SUCCESS em execução do ciclo correspondente.
N03 e N06 possuíam execução ainda em andamento no último snapshot desta redação; não são marcados como VALIDATED até conclusão.
N06 teve um conjunto de erros TypeScript pré-existentes fora do RGO; a verificação RGO foi restringida ao contrato tocado para não mascarar nem reescrever o código ativo. Esses erros permanecem registrados como dívida existente do núcleo.

## 13. Resultado de segurança
Não há afirmação de “sistema perfeito” ou “sem erros”. O que existe é uma implantação com estados explícitos, testes executáveis e evidência de CI onde já concluída.

Pronto para promoção significa pronto para um escopo/versão/tempo definidos; não significa ausência futura de falhas. A evolução permanece habilitada.

## 14. Próximas integrações estruturais exigidas
1. Conectar BugShield real ao envelope RGO, preservando seu schema público original e seus metadados de detecção.
2. Fazer N07 rgo.ingest enviar o envelope a SARA somente quando a fronteira estiver configurada; caso contrário permanecer BLOCKED, nunca PASS.
3. Promover RGO para a cadeia de reauditoria de M01/M03 sem criar uma segunda autoridade.
4. Ligar evidência/histórico ao mecanismo durable já existente de SARA quando a persistência configurada permitir, mantendo UNMEASURABLE quando não houver armazenamento verificável.
5. Usar as PRs históricas (HortaCore, Vagus HTTP, SuperAGI, Mesh/Octacore) como fontes de fusão explícita e não como substitutos silenciosos.

## 15. Regra final de autoaplicação
Qualquer falha encontrada nesta própria infraestrutura gera novo finding e deve passar pelo mesmo ciclo. Nenhum componente do protocolo recebe privilégio de ficar fora da auditoria.
