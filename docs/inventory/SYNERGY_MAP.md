# FASE 0 — SYNERGY MAP DOS 8 NÚCLEOS

Data: 2026-09-30

## Convenções
- **SIM*** = existe caminho/contrato de software; não significa LIVE VERIFIED.
- **NÃO** = não foi encontrado caminho direto confiável no escopo.
- **PENDING** = a ferramenta disponível não permitiu fechar a prova.
- Capacidade "ativa" abaixo significa executável por código quando suas dependências/ambiente existem.
- LIVE VERIFIED continua exigindo URLs reais, credenciais reais, correlação e observação ponta a ponta.

## S1 — MATRIZ 8×8

| De \ Para | N01 | N02 | N03 | N04 | N05 | N06 | N07 | SARA |
|---|---|---|---|---|---|---|---|---|
| N01 | — | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* boundary |
| N02 | SIM* Clareira/Mesh | — | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* HTTP |
| N03 | SIM* Mesh/Clareira | SIM* Gemini bridge | — | SIM* Mesh | SIM* Mesh | SIM* Mesh | PENDING outbound direto | SIM* HTTP |
| N04 | SIM* Mesh | SIM* Mesh | SIM* Mesh | — | SIM* Mesh | SIM* Mesh | SIM* Mesh | PENDING/branch boundary |
| N05 | SIM* Mesh/registration | SIM* Mesh | SIM* Mesh | SIM* Mesh | — | SIM* Mesh | SIM* Mesh | PENDING/branch boundary |
| N06 | SIM* Mesh/Clareira | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | — | SIM* Mesh | PENDING/branch boundary |
| N07 | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | SIM* Mesh | — | SIM* HTTP/Vagus |
| SARA | SIM* federation | SIM* federation | SIM* federation | PENDING | PENDING | PENDING | SIM* N07 boundary | — |

### Leitura da matriz
A maioria das ligações já possui código de transporte e/ou boundary. O gargalo sistêmico não é falta absoluta de canais, mas comprovação de executor, configuração e correlação real. Algumas direções não possuem um peer-client direto no núcleo analisado e por isso permanecem PENDING em vez de serem inferidas.

## S2 — CADEIAS CANÔNICAS

### CADEIA_IA — N03 → N04 → N05
**Estado estrutural:** ATIVA.
- N03 possui peer Mesh para N04.
- N04 possui adapter/handler para peers, incluindo N05.
- N05 possui runtime de inference real.
**Capacidade emergente:** percepção/áudio → tools/artifacts/context → inference/conversation.
**Adormecido:** prova LIVE completa da sequência e configuração de todos os peers/providers.

### CADEIA_EXEC — N01 → N07 → SARA
**Estado estrutural:** ATIVA.
- N01 possui Mesh para N07.
- N07 possui boundary SARA.
- SARA possui runtime regenerativo real.
**Estado RGO:** ADORMECIDO NO MAIN, porque a integração específica RGO/Trinity/MMD continua nas PRs congeladas #60/#54/#24.
**Capacidade emergente:** runtime/transport → orchestration/compute → regeneration/governance.
**Adormecido:** transação RGO/Trinity/Horta ponta a ponta.

### CADEIA_CONV — N02 → N06
**Estado estrutural:** ATIVA.
**Capacidade emergente:** conversação + geração → sessão/contexto/cognitive tools.
**Adormecido:** LIVE end-to-end sem URLs/credenciais/serviços observados.

### CADEIA_UI — N04 → N07 → N01
**Estado estrutural:** ATIVA.
**Capacidade emergente:** interface/tools/artifacts → orchestration → runtime/Mesh/Horta.
**Adormecido:** prova de implantação completa e fluxo real correlacionado.

## S3 — GATILHO DE ELO
A ausência de um elo não deve ser substituída por simulação. Ela deve aparecer como BLOCKED/PENDING. Assim, a capacidade adormecida permanece explicitamente recuperável.

## S4 — PAR NATURAL DE CADA NÚCLEO

| Núcleo | Par natural | Justificativa |
|---|---|---|
| N01 | N07 | runtime/Mesh encontra orchestration/compute |
| N02 | N06 | conversation encontra session/context |
| N03 | N04 | percepção encontra tools/artifacts |
| N04 | N05 | tools/artifacts alimentam inference |
| N05 | N07 | inference encontra policy/orchestration |
| N06 | N02 | contexto encontra diálogo |
| N07 | SARA | orchestration encontra regeneration/governance |
| SARA | N01 | governance/provenance encontra runtime/memory |

## Resultado
O Octacore não aparece como estrela com sete satélites. Há uma malha de capacidades com ownership distribuído. O que falta para transformar esta matriz estrutural em malha operacional é validação executável e configuração/observação de fronteiras, não um segundo cérebro.
