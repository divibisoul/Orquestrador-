# N07 ↔ SARA

## Autoridade

N07 permanece responsável pela orquestração, discovery, roteamento, correlação,
Mesh e recursos de execução. SARA permanece responsável pela regeneração,
auditoria, ética, estratégia, memória, rollback, proveniência e governança do
seu próprio ciclo.

## Configuração

No ambiente do N07:

- SARA_SERVICE_URL=https://<sara-host>
- SARA_SERVICE_TOKEN=<segredo server-side>
- SARA_REQUEST_TIMEOUT=30s

Nenhum segredo deve entrar no repositório.

## Operações registradas

Quando os dois valores de configuração existem, o N07 registra:

- sara.cycle@1.0.0
- sara.audit@1.0.0
- sara.regenerate@1.0.0
- sara.state@1.0.0
- sara.capabilities@1.0.0

A descoberta normal do N07 informa essas operações no inventário.

## Execução

`POST /v1/execute` mantém o contrato N07 existente. Para operações SARA, o
input textual é transportado em metadata.sara_input e o resultado completo do
SARA é preservado em metadata.sara_result_json.

No Mesh `/api/soul-mesh`, o gateway aceita payload estruturado para
capacidades `sara.*` e continua usando o contrato SOUL Mesh 1.1.0, HMAC,
correlationId e anti-replay.

## Falha

Se o serviço SARA não estiver configurado, as operações `sara.*` não são
registradas. Não existe fallback falso.

Quando configurado e indisponível, o N07 retorna a falha real do transporte e
mantém trace/correlation para diagnóstico.
