# AETERNUM — Autoridade funcional integrada ao N07

O N01 é a fonte do grafo AETERNUM existente. Este repositório não copia o runtime e não assume a execução das capacidades de N01.

## Contrato de cooperação

- **Execução:** permanece no `executionOwner` do módulo.
- **Governança:** permanece separada em `governanceAuthority`.
- **Autoridade funcional:** é uma projeção da estrutura de conexões do grafo real.
- **N07:** consome essa evidência para composição/orquestração; não substitui N01.
- **SARA:** continua a autoridade transversal de governança quando declarada.

## Evidência

Fonte exata: `divibisoul/aeternum-core-29@928e7403453bbc8af921a596645e4d9cd9b4cf78`, snapshot gerado pelo resolvedor do N01.

O snapshot é validado no CI do N07. O `source.ref` aponta para o commit imutável do mapa AETERNUM que originou esta projeção. Não é prova de runtime online e não é uma segunda matriz de ownership.

## Relação com NVOD

O NVOD continua como envelope interno de fusão do N01, com mapeamento para o Soul Mesh canônico na fronteira. O N07 apenas consome a projeção de autoridade; não cria outro bus.

## Regra anatômica aplicada

A metáfora da "aorta" é usada somente como modelo estrutural: um módulo com mais conexões funcionais pode ser um âncora/mediador mais central. Isso não revoga ownership nativo nem cria uma hierarquia universal.
