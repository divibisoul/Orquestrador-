# GRF + GRCE Upgrade Wave — 2026-10-06

## Operação

Esta frente é estritamente aditiva. O N07 main existente permanece intocado enquanto a nova camada é preparada em branch isolada.

Objetivos fechados desta onda:

1. Materializar seis upstreams públicos reais por submodule, cada um com revisão fixa.
2. Registrar os candidatos que não puderam ser verificados como UNRESOLVED, sem inventar existência.
3. Adicionar o contrato executável do Golden Rule Framework (GRF).
4. Adicionar o executor Golden Rule Cycle Executor (GRCE), com paralelismo determinístico nas fronteiras SuperGPU/Octacore e preservação de evidência/proveniência.
5. Manter a separação entre órgão, vaso, autoridade e fonte externa.
6. Não reescrever nem absorver as frentes N07 existentes de HortaCore, Clareira, Octacore, SuperGPU, Prefrontal, Mesh ou SARA.

## Estado observado antes da mutação

O N07 main já contém runtime para HortaCore, Clareira, Octacore e SuperGPU. Há PRs abertas separadas para saúde do HortaCore (#128), E2E HortaCore (#125), telemetria/estado do Prefrontal (#127/#135/#138), integridade HMAC do Mesh (#139), capacidade do Octacore (#140) e persistência do backend (#141). Esta frente não altera nenhum desses caminhos.

## Upstreams reais anexados

| Órgão | Upstream | Revisão | Estado |
|---|---|---|---|
| Octacore | lispking/octos | 26a916be14dee49bcc48fbc048947d018c7de6ba | STRUCTURAL_ONLY |
| HortaCore | Vivien83/hora-graph-core | 334f4f82d1c7525d29130ced9e6199e3633c0f61 | STRUCTURAL_ONLY |
| Clareira | mycelium-io/mycelium | 26156b5e9f169e74c73848908ad5e3bad76cd6fa | STRUCTURAL_ONLY |
| PrefrontalNeocortex | PrimeIntellect-ai/prime-agent | 7a52276cb17310f331f1075f28fa5cf9c4ae0a0a | STRUCTURAL_ONLY |
| SuperGPU | SuperInstance/cuda-oxide | 1ad9bfc5241322ceeccb3f9182f2a188a648bd08 | STRUCTURAL_ONLY |
| Orbitador | SuperInstance/openmind-conductor | 18c5498a4175ed0b20bff03bc3cf2ed3233be014 | STRUCTURAL_ONLY |

STRUCTURAL_ONLY significa: fonte materializada e pinada; a existência da fonte não autoriza declarar seu runtime como adaptado, conectado ou online.

## GRF

O pacote grf centraliza:

- estados epistêmicos;
- evidência hasheada;
- proveniência;
- falha → caracterização → oposição → análise → integração → transformação;
- invariantes I1–I18;
- validação de monotonicidade e hashes.

A implementação não duplica ARA, ETR, ITR, MMD, ERU ou RGO. Ela define os contratos para que essas autoridades existentes possam participar do ciclo por adaptadores.

## GRCE

O pacote grce executa a sequência:

DETECT → EVIDENCE → CHARACTERIZE → DUALIZE/ETR GATE → ANALYZE → MMD → ARA → ETR → ITR → FORM → VALIDATE → HORTA/VAGUS/MESH/ERU.

SuperGPU e Octacore permanecem como produtores paralelos distintos. A fusão é feita por contrato; nenhuma autoridade externa é promovida a root.

Falhas de preflight, contexto, dualização, monotonicidade, validação ou rollback retornam estado PRESERVED/UNRESOLVED e mantêm a evidência disponível.

## Limite de ativação

Nenhuma integração externa desta onda é marcada como LIVE. O runtime principal continua sendo o sistema nativo já presente no SOUL. A ativação funcional futura exige adaptador explícito, teste determinístico e, para rede/serviço, prova E2E autenticada.
