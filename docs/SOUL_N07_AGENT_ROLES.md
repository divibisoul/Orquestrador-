# N07 — agentes locais do Orquestrador

N07 é o único núcleo que coordena a federação. Os agentes abaixo permanecem dentro do N07 e não substituem agentes dos N01–N06.

| Agente | Responsabilidade |
|---|---|
| n07.discovery | Descobrir peers/capabilities e resolver o proprietário executável de uma capability. |
| n07.router | Selecionar peer/rota usando capability, saúde, latência e carga. |
| n07.executor | Executar localmente ou disparar execução SuperGPU/federada autorizada. |
| n07.composer | Compor capabilities em fusões dinâmicas somente quando o contrato permitir. |
| n07.storage | Armazenar conteúdo endereçado por conteúdo e consultar status de armazenamento. |
| n07.validator | Validar resultado, correlationId e versão de contrato antes de aceitar o resultado. |
| n07.observer | Observar saúde, métricas e tracing da federação. |
| n07.jev | Produzir decisões tipadas/probabilidades/confiança através de `jev.systemone@1.0.0`. |
| n07.cooperation | Executar handshake, exchange e health do plano cooperativo, preservando correlação e usando o Mesh canônico. |

**Limite:** N07 orquestra; não absorve os motores nativos dos outros núcleos. Gemini continua provido por N02, áudio por N03, ferramentas/documentos por N04, inferência por N05 e cognição por N06. SARA continua transversal.
