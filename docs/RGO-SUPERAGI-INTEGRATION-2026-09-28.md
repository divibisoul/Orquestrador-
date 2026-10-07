# RGO → SUPER AGI — Contrato de Integração
Data: 2026-09-28

## Objetivo
Implantar a Regra de Ouro como infraestrutura executável de auditoria/transmutação em oito núcleos, preservando autoridades e artefatos existentes.

## Lei operacional
F → N(F) → D(F) → C(D(F)) → I → V → H.

Entrada é material de auditoria, nunca verdade presumida. Hipótese, requisito, implementação, teste e validação permanecem epistemicamente distintos.

## Regras
1. Nada é apagado, silenciado ou substituído por cópia paralela.
2. Soul Mesh 1.1.0 continua o transporte inter-núcleos canônico.
3. Vagus/VagusNerveBus é barramento de eventos/instrumentação; não substitui o Mesh.
4. HortaCore é memória/substrato; não é autoridade de governança.
5. Ausência de evidência nunca produz PASS.
6. Dual sem correction_boundary.required_property = UNRESOLVED.
7. PromotionGate executa contrato/política; não inventa critérios.
8. UtilityPolicy e StoppingPolicy não recebem limiares inventados pelo runtime.
9. Capacidade possui estados separados: DERIVED, IMPLEMENTED, TESTED, VALIDATED, PROMOTABLE, PROMOTED.
10. Toda integração deve ser reauditável e preservar proveniência.

## Distribuição
| Núcleo | Papel |
|---|---|
| N01 | captura local, HortaCore, nervoVago e saída Mesh |
| N02 | adaptador RGO sobre Mesh |
| N03 | adaptador RGO sobre Mesh |
| N04 | adaptador RGO sobre Mesh |
| N05 | adaptador RGO sobre Mesh |
| N06 | adaptador RGO sobre Mesh |
| N07 | ingestão/orquestração e fronteira SARA |
| G0/SARA | autoridade de governança/regeneração/proveniência |

## Estado observado antes desta branch
N01 possui HortaCore em memória e fachada nervoVago sobre EventBus real.
SARA possui VagusNerveBus in-process, ProvenanceTracker e ConnectedRuntime.
N07 possui SARAProxy e já envia envelopes Vagus para /v1/vagus, mas a implementação desse endpoint está em PR aberto de SARA, não em main.
N07 HortaCore histórico está em PR #47, não em main.
SuperAGI está em PR #49 e não é afirmado como execução live.

## Segurança de branches
A implantação é isolada em rgo-integration-2026-09-28, criada de cada main. PRs já existentes não são reescritos nem absorvidos automaticamente. Nenhum merge automático é feito.

## Critério de evidência
Código novo recebe testes executáveis. PASS só pode ser declarado após execução real do teste ou CI. Dependências externas continuam BLOCKED ou UNMEASURABLE até evidência correspondente.
