package security

import "testing"

func TestValidateRejectsInvalidInputs(t *testing.T) {
	p := Policy{AllowExternalEffects: true, MaxCost: 10, MinConfidence: 0.5}
	if err := Validate(p, -1, 0.8); err == nil { t.Fatal("expected invalid cost") }
	if err := Validate(p, 1, 1.5); err == nil { t.Fatal("expected invalid confidence") }
	if err := Validate(p, 11, 0.8); err == nil { t.Fatal("expected cost rejection") }
	if err := Validate(p, 1, 0.4); err == nil { t.Fatal("expected confidence rejection") }
}

func TestAuthorizeRequiresPrincipalAndAction(t *testing.T) {
	if err := Authorize("", "execute"); err == nil { t.Fatal("expected authentication requirement") }
	if err := Authorize("node-1", ""); err == nil { t.Fatal("expected action requirement") }
	if err := Authorize("node-1", "execute"); err != nil { t.Fatal(err) }
}
