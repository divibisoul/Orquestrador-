# N07 → SARA → N01 — RGO Trinity/MMD

N07 permanece autoridade de orquestração.

rgo.trinity.process@1.0.0 envia o finding para SARA, recebe a sequência de StageEnvelope e encaminha cada etapa ao N01 por Soul Mesh com rgo.hortacore.store.

A resposta só recebe VALIDATED_WITH_HORTA quando todas as etapas têm resposta comprovada do HortaCore. Falha de infraestrutura gera PARTIAL_BLOCKED_HORTA e nunca PASS sintético.

O transporte entre núcleos continua separado do Vagus.