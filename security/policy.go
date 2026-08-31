package security

import "errors"
type Policy struct{AllowExternalEffects bool; MaxCost float64; MinConfidence float64}
func Validate(p Policy,cost,confidence float64)error{if cost>p.MaxCost{return errors.New("cost exceeds policy")};if confidence<p.MinConfidence{return errors.New("confidence below policy")};if !p.AllowExternalEffects{return errors.New("external effects disabled")};return nil}
