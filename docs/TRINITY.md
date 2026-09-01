# Tríplice Fundamental

**Neo-Córtex (PFC) ↔ Neural Fabric ↔ Transcendental Compute Engine**

A Tríplice é uma camada experimental, determinística e isolada para decisão executiva, roteamento adaptativo e execução computacional simulada.

## Metáforas computacionais

- **Miller** → working memory e gating: capacidade limitada e seleção de informação relevante.
- **Gershman** → estado: manutenção explícita de contexto operacional.
- **Botvinick** → meta-RL: política leve que escolhe precisão/paralelismo e atualiza preferências por recompensa.
- **Cohen** → conflito: sinal composto de custo, incerteza/histórico e ambiguidade.
- **Numenta** → modelos concorrentes: múltiplas rotas/modelos avaliados por regras e pesos, sem alegar equivalência biológica.

## Fluxo

`Task → Workload → PFC.Evaluate → Fabric.Route → Compute.Execute → Fabric.Feedback → PFC.GateWorkingMemory`

Se o PFC rejeitar ou o risco exceder o limiar, `fallback_mode` controla `legacy`, `retry` ou `skip`.

## Flags

Por padrão todas as flags estão desligadas. Isso preserva o caminho legado.

```yaml
trinity:
  pfc_enabled: false
  fabric_enabled: false
  compute_enabled: false
  risk_threshold: 0.75
  working_memory_limit: 16
  fallback_mode: legacy
```

Para ativar a Tríplice, todas as três flags precisam ser `true` e as dependências devem estar presentes.

## Compute

O Transcendental Compute Engine é **simulação**, não um driver CUDA/ROCm/NPU. As heurísticas selecionam modelos virtuais: `blackwell`, `vera_rubin`, `mi400`, `atlas` e `trillium`.

Heurísticas:

1. FP64 → blackwell
2. MemoryNeeded > 400 → mi400
3. MatrixSize > 4096 → vera_rubin
4. MatrixSize < 1024 → trillium
5. BatchSize > 64 → blackwell
6. default → blackwell

## Limitações

A Tríplice não afirma reproduzir neurobiologia. Os nomes e referências são metáforas de engenharia. O compute atual é simulado e não representa desempenho de hardware real. O aprendizado é um mecanismo de pesos simples, não treinamento de rede neural.
