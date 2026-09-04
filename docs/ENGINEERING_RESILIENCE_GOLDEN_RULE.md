# Engenharia de Resiliência Generativa — Regra de Ouro do SOUL N07

## Finalidade

Toda falha estrutural deve atravessar três camadas: reação, correção e geração de conhecimento operacional. Uma falha nunca é convertida em sucesso artificial.

## Camada Reativa

`resilience.Engine` aplica:

- retry limitado com backoff exponencial;
- circuit breaker com estados `CLOSED`, `OPEN` e `HALF_OPEN`;
- fallback explícito para compute quando o caminho primário falha;
- preservação de erro e stack trace;
- idempotência por desenho: uma execução de compute não muta o estado do runtime antes de o resultado ser aceito.

## Camada Corretiva

Cada falha possui assinatura estável, severidade, operação, tentativa, estado e camada. O `FaultStore` mantém o registro forense; o chamador pode classificar a causa com FMEA/5 Whys sem substituir o evento bruto.

Estados relevantes: `HEALTHY`, `DEGRADED`, `FAILED`, `NOT_CONFIGURED`.

A política de circuito impede que uma falha persistente continue degradando o peer indefinidamente. O fallback não promove o sistema para `HEALTHY`: sucesso no fallback resulta em `DEGRADED`.

## Camada Generativa

A mesma falha produz:

1. métrica Prometheus `soul_resilience_faults_total` com assinatura e estado;
2. histórico estruturado via `FaultStore`;
3. testes de caos controlado em `golden_rule_test.go`;
4. suporte a runbooks automatizados, protegido por `RemediationPolicy`.

A remediação automática nasce desarmada por padrão. Quando armada, só executa runbooks dentro do limite de severidade e após o número mínimo configurado de ocorrências. Isso controla blast radius e impede que um incidente crítico gere uma alteração autônoma em produção sem autorização operacional.

## Loop de Aprendizado

`falha -> assinatura + stack -> registro forense -> teste de não-regressão/caos -> métrica -> decisão de remediação -> novo estado observável`.

A camada não altera contratos Mesh automaticamente. Alterações de `soul-mesh/1@1.1.0`, Protobuf/OpenAPI ou topologia permanecem responsabilidade das autoridades canônicas do SOUL e devem entrar por mudança versionada e validada.

## Limite de hardware

O runtime SuperGPU possui fallback CPU real para workloads suportados quando o caminho acelerado falha. A existência de `nvidia-smi` ou `rocminfo` não é tratada como prova de execução CUDA/ROCm. VRAM e temperatura só podem ser reportadas como medidas quando o backend expuser esses dados reais.

## Evidência

O gate do N07 deve executar `go test ./resilience` e `go test -race ./resilience`. A promoção para `ONLINE` continua condicionada aos gates de arquitetura, testes, E2E e evidência de hardware reais.
