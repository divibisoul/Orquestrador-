package aeternum
import "testing"
func TestValidateCatalogContainsExpectedAutonomousSurface(t *testing.T){ if err:=ValidateCatalog(); err!=nil { t.Fatal(err) } }
