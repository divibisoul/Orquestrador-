# FASE 0 — RELATÓRIO GERAL
Data: 2026-09-30
Escopo: N01–N07 + SARA.

## A — Auditoria
Os relatórios por núcleo estão em docs/forensics/<repo>/REPORT.md. O marco-zero da Clareira foi definido pela base da respectiva PR de Fase1. Onde o conector não permitiu git log -n 500, tags ou diff recursivo integral, o estado foi marcado NOT MEASURED/PENDING; nenhuma conclusão foi inventada.

## B — Inventário
Os inventários estão em docs/inventory/<repo>/TOOLS.md. Eles registram processadores, endpoints, funções públicas principais, eventos, externos, dependências inter-núcleo, ferramentas adormecidas e expansão por conexão.

## C — N01
V1 EventBus: OK, único barramento canônico em src/core/EventBus.ts; nervoVago/AeternumEventBus são fachadas.
V2 ModuleRegistry: OK.
V3 Clareira: 9 nós preservados: NC-001, NP-001..003, NS-001..005.
V4 neural: OK estrutural; simulações legadas ficam PENDING.
V5 Mesh/Horta: preservados; RGO Trinity continua BRANCH_ONLY.
V6 App.tsx: preservado.
V7 configs: preservados no escopo observado.

PR #69 adiciona um gate executável de continuidade do Core. Head 6486fae49b759ee9ee4c4c631c228485b8374944. CI atual não foi medido pelo conector; portanto não há declaração de PASS.

## ESTADO DOS 8 NÚCLEOS

| Núcleo | Processadores por código | Endpoints LIVE VERIFIED | NAME_ONLY/condicional | Soterrado comprovado | Deletado comprovado |
|---|---|---:|---|---:|---:|
| N01 | ~17 | 0 | Gemini/SARA/peer | 0 crítico | 0 |
| N02 | ~6 | 0 | capability sem executor/provider | 0 crítico observado | não comprovado |
| N03 | ~5 | 0 | Gemini/SARA/peer | 0 crítico observado | não comprovado |
| N04 | ~6 | 0 | sessão/context tools | 0 crítico observado | não comprovado |
| N05 | ~5 | 0 | provider/peer | 0 crítico observado | não comprovado |
| N06 | ~6 | 0 | sessão/context tools | 0 crítico observado | não comprovado |
| N07 | ~7 famílias | 0 | SARA/Jev | 0 crítico observado | não comprovado |
| SARA | ~13 | 0 | broker/crawlers/oracles | 0 crítico observado | não comprovado |

0 LIVE VERIFIED é deliberado: código executável ≠ deployment real.

## PRs congeladas
- SARA #24 — RGO/Trinity/MMD
- N01 #60 — HortaCore bridge
- N07 #54 — fail-closed RGO Trinity
Não foram mergeadas.

## Frentes cooperativas
N01 #63/#65/#67/#68; N02 Gemini/ATLAS; N03 real Gemini + correlation; N04 capability/tool compatibility + ATLAS; N05 real inference + ATLAS; N06 context + ATLAS; N07 cognitive/primordial/NeuralForge/ASC; SARA Bayesian/OctaCore/Vagus. Cada frente permanece em seu lineage e não é duplicada.

## Capacidades adormecidas
RGO Trinity no MAIN; Mesh LIVE entre deployments; Gemini sem credencial; SARA sem URL/token; Jev sem JEV_API_KEY; contextual tools sem user session; N02 NeuralForge/ASC sem executor comprovado; SARA external infrastructure sem configuração; GEM-Health sem fonte wearable.

## GAPS
Sem prova LIVE ponta a ponta da malha. Sem inventário source-backed dos 70/72 módulos nesta rodada. Tags e diffs integrais de todas as refs não foram mensuráveis pelo conector. Algumas direções inter-núcleo continuam PENDING quando não há outbound client local comprovado.

## SINERGIA
CADeIA_IA N03→N04→N05: estruturalmente ativa; LIVE PENDING.
CADEIA_EXEC N01→N07→SARA: estruturalmente ativa; RGO específico adormecido no MAIN.
CADEIA_CONV N02→N06: estruturalmente ativa; LIVE PENDING.
CADEIA_UI N04→N07→N01: estruturalmente ativa; LIVE PENDING.

Mapa NxN completo: docs/inventory/SYNERGY_MAP.md.

## ORDEM DE MERGE RGO
SARA #24 → N01 #60 → N07 #54.
Depois: smoke correlacionado ponta a ponta e comprovação de final_status=VALIDATED + persistência HortaCore de todas as etapas.

## EPISTEMOLOGIA
DECLARED ≠ EXECUTABLE ≠ RUNNING ≠ VERIFIED ≠ LIVE VERIFIED.
Guard não cria evidência. CI não cria deployment. Health não prova transação.

## PERGUNTAS PARA DIEGO
1. Qual é o SPEC canônico de neural.forward e do fluxo de requisição?
2. Onde está o artefato original dos 70/72 módulos?
3. Qual política oficial de versionamento entre os oito núcleos?
4. Qual implementação é autoridade quando duas capacidades históricas se sobrepõem?
5. Qual é o fluxo canônico usuário → percepção/conversa/tools/context → N07 → N01/SARA → resposta?
6. O que define oficialmente uma conexão como ativa além de existência de código/CI?
7. Quais URLs reais poderão ser usadas futuramente no smoke LIVE? Não envie secrets pelo chat.
