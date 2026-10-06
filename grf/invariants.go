package grf

import (
	"errors"
	"fmt"
)

var CanonicalInvariants = []Invariant{
	{ID:"I1", Description:"Não-eliminação: conteúdo existente é preservado; preservar não implica ativar."},
	{ID:"I2", Description:"Monotonicidade de conteúdo: o novo estado não pode ser menor que o anterior."},
	{ID:"I3", Description:"Proveniência completa: parent_hash, input_hash, output_hash e sequence_index são obrigatórios."},
	{ID:"I4", Description:"Dual explícito: uma oposição só existe quando a propriedade necessária estiver declarada."},
	{ID:"I5", Description:"Rollback é segurança, não evolução."},
	{ID:"I6", Description:"Toda transformação precisa de teste antes da incorporação."},
	{ID:"I7", Description:"Toda nova capacidade aponta causalmente para sua falha de origem."},
	{ID:"I8", Description:"Implementação defeituosa permanece preservada e bloqueada; nunca apagada."},
	{ID:"I9", Description:"Nenhuma operação executa sem contexto."},
	{ID:"I10", Description:"Todo ciclo realimenta ERU, RGO, VagusBus e Mesh."},
	{ID:"I11", Description:"Os componentes regenerativos não são intercambiáveis."},
	{ID:"I12", Description:"A topologia é circular; nenhum componente é raiz absoluto."},
	{ID:"I13", Description:"Órgão e vaso são distintos; transporte não é estado."},
	{ID:"I14", Description:"Comunicação é adjacente; não há salto silencioso de camada."},
	{ID:"I15", Description:"VagusBus transporta sinal; HortaCore transporta estado."},
	{ID:"I16", Description:"Paralelismo produz resultados mergeáveis com hash próprio."},
	{ID:"I17", Description:"O gate ético ocorre antes da derivação/uso do dual."},
	{ID:"I18", Description:"Transformação segue ARA → ETR → ITR sem atalho."},
}

func ValidateInvariantSet(got []Invariant) error {
	seen := make(map[string]bool, len(got))
	for _, inv := range got {
		if seen[inv.ID] {
			return fmt.Errorf("GRF_DUPLICATE_INVARIANT:%s", inv.ID)
		}
		seen[inv.ID] = true
	}
	for _, required := range CanonicalInvariants {
		if !seen[required.ID] {
			return fmt.Errorf("GRF_MISSING_INVARIANT:%s", required.ID)
		}
	}
	return nil
}

func ValidateContextAndProvenance(before State, ctx Context, after State, p Provenance) error {
	if err := ctx.Validate(); err != nil {
		return err
	}
	if before.Epistemic == "" || after.Epistemic == "" {
		return errors.New("GRF_EPISTEMIC_STATE_REQUIRED")
	}
	if after.Size() < before.Size() {
		return fmt.Errorf("GRF_MONOTONICITY_VIOLATION:before=%d after=%d", before.Size(), after.Size())
	}
	if p.ParentHash == "" || p.InputHash == "" || p.OutputHash == "" || p.SequenceIndex == 0 {
		return errors.New("GRF_PROVENANCE_INCOMPLETE")
	}
	if p.ParentHash != before.Hash() {
		return errors.New("GRF_PARENT_HASH_MISMATCH")
	}
	if p.InputHash != hashState(before) || p.OutputHash != hashState(after) {
		return errors.New("GRF_STATE_HASH_MISMATCH")
	}
	return nil
}

func hashState(s State) string {
	return s.Hash()
}

func PreserveOnFailure(before State) State {
	next := before.Clone()
	next.Epistemic = EpistemicPreserved
	return next
}
