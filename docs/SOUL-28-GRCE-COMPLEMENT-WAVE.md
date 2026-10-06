# SOUL-28 — GRCE complement wave

Parent graph: SOUL-25-capability-fabric  
Parent file SHA-256: `sha256:15f70a247216327f51f064c115a9d23a8f27ac122def19604a0b7f832d1b1313`

## Preservação

A estrutura SOUL-25 permanece preservada: 25 nós e 144 arestas originais continuam presentes. SOUL-28 acrescenta exatamente 3 nós e 12 arestas, totalizando 28 nós e 156 arestas. A cópia `_preserved/soul-25-capability-fabric.json` é verificada contra o `parent_hash` durante o gate.

## Complementos

| Complemento | Função no GRCE | Estado epistemológico |
|---|---|---|
| `bijux-dag-runtime` | execution-kernel, replay e proveniência | PROJECTED |
| `ouro-loop` | self-healing limitado e verificação | PROJECTED |
| `recurs` | memória experiencial e localização de falhas | PROJECTED |

Os três são submodules pinados e possuem adaptadores `GoldenRuleParticipant`. A presença do gitlink não é prova de execução.

## Fronteira de execução externa

O contrato de transporte é definido pelo SOUL no adaptador N07; ele não é atribuído automaticamente ao upstream. Um runtime externo configurado deve devolver:

- identidade `Source` e `Revision` exatamente iguais ao pin esperado;
- estado em um dos estados epistemológicos canônicos;
- evidência com `ID`, `ContextHash`, `Source` e hash verificável;
- proveniência com `ParentHash`, `InputHash`, `OutputHash`, `SequenceIndex` e cadeia não vazia;
- estado serializado em base64, cujo hash final corresponda ao `OutputHash`.

Ausência de comando, falha de execução, JSON inválido, estado inválido, divergência de identidade, hash ou proveniência preservam a entrada, registram evidência e mantêm o complemento PROJECTED/BLOCKED. Nenhuma dessas condições pode produzir REAL/ACTIVE por declaração.

## Topologia

O gate verifica não apenas contagem, mas também os 12 elos esperados, seus modos, estado PROJECTED e grau 4 de cada complemento. Duplicidade de arestas e divergência do hash do pai são erros de integridade.

## Limite de autoridade

N01–N07, SARA e JEV continuam com sua autoridade nativa. N07 continua sendo o control plane/federação. Nenhum complemento externo cria um segundo Mesh, substitui HortaCore, substitui Clareira, altera a autoridade de SARA ou promove um novo núcleo SOUL.

## Evidência atual

**REAL:** pins Git verificáveis, topologia 28/156, hash do pai, contratos/adaptadores e gates executados com sucesso nos heads já validados.  
**PROJECTED:** execução dos três runtimes externos até existir comando configurado e resposta observável validada pelo contrato.  
**BLOCKED:** ambientes sem o comando/runtime externo configurado ou com metadados/proveniência inválidos.  
**PRESERVED:** SOUL-25 original, histórico Git e falhas encontradas durante a integração.
