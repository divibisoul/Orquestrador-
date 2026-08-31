package security

import (
	"crypto/tls"
	"testing"
)

func TestMTLSValidation(t *testing.T) {
	if err := (MTLSConfig{}).Validate(); err == nil { t.Fatal("expected missing certificate error") }
	if err := (MTLSConfig{CertFile:"c",KeyFile:"k",CAFile:"ca",MinVersion:tls.VersionTLS12}).Validate(); err == nil { t.Fatal("expected TLS 1.2 rejection") }
	if err := (MTLSConfig{CertFile:"c",KeyFile:"k",CAFile:"ca",SPIFFETrustDomain:"bad domain"}).Validate(); err == nil { t.Fatal("expected SPIFFE trust-domain validation error") }
	if err := (MTLSConfig{CertFile:"c",KeyFile:"k",CAFile:"ca",SPIFFETrustDomain:"soul.local"}).Validate(); err != nil { t.Fatalf("valid SPIFFE trust domain rejected: %v", err) }
}
