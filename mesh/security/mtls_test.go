package security

import (
	"crypto/tls"
	"testing"
)

func TestMTLSValidation(t *testing.T) {
	if err := (MTLSConfig{}).Validate(); err == nil { t.Fatal("expected missing certificate error") }
	if err := (MTLSConfig{CertFile:"c",KeyFile:"k",CAFile:"ca",MinVersion:tls.VersionTLS12}).Validate(); err == nil { t.Fatal("expected TLS 1.2 rejection") }
}
