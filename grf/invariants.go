package grf

import (
	"errors"
	"sort"
)

type Invariant struct {
	ID string
	Description string
}

type InvariantSet struct { Invariants []Invariant }

var canonicalInvariants = []Invariant{
	{ID:"I1",Description:"Não-eliminação; preservar não implica ativar."},
	{ID:"I2",Description:"Monotonicidade de conteúdo; regeneração que reduz conteúdo aborta."},
	{ID:"I3",Description:"Completude de proveniência."},
	{ID:"I4",Description:"Derivação explícita de dual."},
	{ID:"I5",Description:"Rollback é segurança, não evolução."},
	{ID:"I6",Description:"Toda transformação precisa de teste antes de incorporação."},
	{ID:"I7",Description:"Toda nova capacidade aponta causalmente para a falha que a originou."},
	{ID:"I8",Description:"Preservar e ativar são operações separadas."},
	{ID:"I9",Description:"Nenhuma operação roda sem contexto C."},
	{ID:"I10",Description:"Todo ciclo realimenta ERU, RGO, VagusBus e Mesh."},
	{ID:"I11",Description:"Componentes regenerativos não são intercambiáveis."},
	{ID:"I12",Description:"Topologia circular; nenhum componente é raiz ou topo."},
	{ID:"I13",Description:"HortaCore e VagusBus transportam estado/sinal; não são o estado."},
	{ID:"I14",Description:"Comunicação entre camadas deve respeitar adjacência."},
	{ID:"I15",Description:"VagusBus = sinal; HortaCore = estado."},
	{ID:"I16",Description:"SuperGPU e Octacore produzem resultados mergeáveis com hash próprio."},
	{ID:"I17",Description:"Gate ético antes de dual."},
	{ID:"I18",Description:"Tripla barreira ARA → ETR → ITR, sem atalho."},
}

func CanonicalInvariantSet() InvariantSet {
	out := make([]Invariant, len(canonicalInvariants))
	copy(out, canonicalInvariants)
	return InvariantSet{Invariants: out}
}

func ValidateInvariantSet(set InvariantSet) error {
	seen := map[string]bool{}
	for _, inv := range set.Invariants {
		if inv.ID == "" || inv.Description == "" { return errors.New("GRF invariant entry is incomplete") }
		seen[inv.ID] = true
	}
	for _, required := range canonicalInvariants {
		if !seen[required.ID] { return errors.New("missing invariant " + required.ID) }
	}
	return nil
}

func InvariantIDs(set InvariantSet) []string {
	out := make([]string,0,len(set.Invariants))
	for _, inv := range set.Invariants { out=append(out,inv.ID) }
	sort.Strings(out)
	return out
}
