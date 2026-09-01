# N07 — Blueprint Real de Execução

## 1. Contrato de engenharia
N07 é o Orquestrador. Ele coordena os núcleos e conecta a rede neural, o módulo de decisão pré-frontal e a camada de computação por interfaces executáveis.

As funções deste repositório devem representar comportamento executável. Não são permitidos placeholders, funções vazias, resultados fabricados, estados de sucesso sem execução ou declarações de capacidade que o backend não possui.

## 1.1 Diretiva operacional contínua
Toda intervenção no GitHub deve começar pela leitura do estado real da branch/HEAD, PRs concorrentes, mudanças recentes e gates. Um erro ou bloqueio não encerra a atividade: deve ser corrigido, pesquisado ou contornado por alternativa técnica válida antes de retornar à tarefa original.

Após cada escrita, o arquivo deve ser lido novamente e o commit/sha verificado. Não se considera uma alteração concluída por uma chamada de ferramenta; a evidência executável é obrigatória quando aplicável.

Branches concorrentes são fontes de trabalho, não versões a serem sobrescritas. Antes de alterar uma área, localizar a implementação canônica e comparar os SHAs relevantes. Duplicações devem ser consolidadas por responsabilidade e contratos.

Tempo também é sinal operacional: enquanto um gate externo estiver processando, outras frentes independentes continuam. O dashboard/gráfico deve acompanhar mudanças materiais e nunca ser abandonado como indicador de escopo, regressão e trabalho restante.

## 2. Estados de uma capacidade
Uma capacidade somente pode ser considerada operacional quando existe o caminho completo:

`request → validação → roteamento → processamento → execução → resultado → erro/cancelamento → liberação de recursos`

Estados conceituais usados na auditoria externa permanecem fora do produto; o código deve provar o comportamento diretamente por testes executáveis.

## 3. As 40 funções públicas
### Orquestrador
`New`, `Register`, `Route`, `Submit`, `Execute`, `Cancel`, `Status`, `Health`, `Stats`, `Shutdown`.

### Rede Neural
`New`, `AddEdge`, `RemoveEdge`, `Activate`, `Forward`, `Learn`, `Normalize`, `Attention`, `Backprop`, `Health`.

### Neocórtex Pré-frontal
`New`, `Evaluate`, `Plan`, `Prioritize`, `Inhibit`, `Select`, `ValidateAction`, `Commit`, `Recall`, `Health`.

### SuperGPU
`New`, `Discover`, `Select`, `Reserve`, `Release`, `Execute`, `Batch`, `MemoryStats`, `Health`, `Shutdown`.

## 4. Invariantes obrigatórios
- Entrada numérica não pode conter NaN ou infinito.
- Índices e dimensões devem ser validados antes do processamento.
- Toda operação deve respeitar `context.Context` quando houver trabalho potencialmente cancelável.
- `trace_id` identifica uma execução e não pode ser reutilizado enquanto estiver ativa.
- Recursos computacionais reservados devem ser liberados em sucesso, erro e cancelamento.
- Um dispositivo preferencial explicitamente solicitado não pode ser trocado silenciosamente por outro.
- Um backend não pode executar um dispositivo que não declara suportar.
- Descobrir uma ferramenta de driver não prova capacidade de execução de kernel.
- Falha de backend deve retornar erro explícito.
- Shutdown deve impedir novas execuções e cancelar as ativas.
- Um runtime computacional recém-construído deve possuir discovery suficiente para iniciar o caminho padrão ou declarar degradação explicitamente.

## 5. Pipeline cognitivo
`cognitive.execute` implementa o caminho N07:

1. validar a mensagem;
2. executar forward da rede neural;
3. derivar uma métrica de utilidade do resultado;
4. submeter a decisão ao módulo pré-frontal;
5. validar risco, custo e limiar;
6. registrar a decisão;
7. selecionar o backend computacional disponível;
8. reservar o dispositivo pelo `trace_id`;
9. executar a operação;
10. liberar o dispositivo e devolver o resultado com rastreabilidade.

## 6. SuperGPU e hardware real
A camada `supergpu` é uma abstração de execução de hardware. O backend padrão executa operações matemáticas no CPU de forma real. NVIDIA e AMD só são marcadas como executáveis quando um backend compatível declara suporte ao dispositivo detectado.

Isso evita confundir detecção de hardware com execução acelerada. Para execução nativa CUDA/HIP, o backend correspondente deve implementar a interface `Backend` e declarar suas capacidades.

## 7. Compatibilidade com N01–N06 e futura fusão com ferramentas/funções
A integração externa deve respeitar o contrato canônico adotado pelo sistema: versão de contrato, discovery, health, circuit breaker, processors únicos e rastreabilidade. O N07 não deve duplicar responsabilidade de outro núcleo nem alterar silenciosamente contratos existentes.

Antes da fusão final com N01 e N06, mapear entradas, saídas, eventos, capacidades, ferramentas e funções de todos os seis núcleos e do control plane. A fusão deve privilegiar contratos compartilhados, capability negotiation e adapters sobre cópia de implementações. Funções e ferramentas equivalentes devem possuir um único owner canônico; capacidades complementares podem ser compostas em novas funções quando houver ganho verificável.

A validação de integração deve provar `request → peer → runtime → execução → response`, e não apenas a existência de endpoints ou tipos.

## 8. Regra de manutenção
Toda área identificada como incompleta, inativa, inconsistente, insegura ou desconectada deve ser corrigida no código e acompanhada de teste executável. Uma descrição de intenção nunca substitui uma implementação funcional.

N07 deve continuar sendo auditado e elevado em paralelo às demais frentes. O fato de sua fusão final ocorrer depois não autoriza deixá-lo tecnicamente atrasado agora.
