package verifier

import (
	"context"
	"errors"
	"strings"

	"github.com/divibisoul/Orquestrador-/core/superagi"
)

// Verifier delegates checks to the canonical SuperAGI runtime.
type Verifier struct { Runtime *superagi.Runtime }

func New(r *superagi.Runtime) *Verifier {
	if r == nil { r = superagi.NewRuntime() }
	return &Verifier{Runtime:r}
}

func (v *Verifier) Fact(ctx context.Context, claim string) (bool, float64, error) {
	if v == nil || v.Runtime == nil { return false, 0, errors.New("verifier runtime unavailable") }
	if strings.TrimSpace(claim)=="" { return false, 0, errors.New("claim required") }
	ok, confidence := v.Runtime.VerifyFact(ctx, strings.TrimSpace(claim))
	return ok, confidence, nil
}
func (v *Verifier) Safety(ctx context.Context, text string) (bool, float64, error) {
	if v == nil || v.Runtime == nil { return false, 0, errors.New("verifier runtime unavailable") }
	if ctx == nil { return false, 0, errors.New("nil context") }
	if err := ctx.Err(); err != nil { return false, 0, err }
	ok, confidence := v.Runtime.VerifySafety(ctx, text)
	return ok, confidence, nil
}
func (v *Verifier) Coherence(ctx context.Context, text string) (bool, float64, error) {
	if v == nil || v.Runtime == nil { return false, 0, errors.New("verifier runtime unavailable") }
	if ctx == nil { return false, 0, errors.New("nil context") }
	if err := ctx.Err(); err != nil { return false, 0, err }
	ok, confidence := v.Runtime.VerifyCoherence(ctx, text)
	return ok, confidence, nil
}
func (v *Verifier) Code(ctx context.Context, code string) (bool, float64, error) {
	if v == nil || v.Runtime == nil { return false, 0, errors.New("verifier runtime unavailable") }
	if ctx == nil { return false, 0, errors.New("nil context") }
	if err := ctx.Err(); err != nil { return false, 0, err }
	ok, confidence := v.Runtime.VerifyCode(ctx, code)
	return ok, confidence, nil
}
