package security

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
)

type MTLSConfig struct {
	CertFile       string
	KeyFile        string
	CAFile         string
	MinVersion     uint16
	ServerName     string
	RequireClient  bool
}

func (c MTLSConfig) Validate() error {
	if c.CertFile == "" || c.KeyFile == "" || c.CAFile == "" {
		return errors.New("mTLS certificate, key and CA files are required")
	}
	if c.MinVersion == 0 {
		return nil
	}
	if c.MinVersion < tls.VersionTLS13 {
		return errors.New("TLS 1.3 or newer is required")
	}
	return nil
}

func LoadMTLSConfig(c MTLSConfig) (tls.Certificate, *x509.CertPool, error) {
	if err := c.Validate(); err != nil {
		return tls.Certificate{}, nil, err
	}
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	caPEM, err := os.ReadFile(c.CAFile)
	if err != nil {
		return tls.Certificate{}, nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return tls.Certificate{}, nil, errors.New("CA file contains no valid certificates")
	}
	return cert, pool, nil
}

func (c MTLSConfig) ServerTLS() (*tls.Config, error) {
	cert, pool, err := LoadMTLSConfig(c)
	if err != nil { return nil, err }
	min := c.MinVersion
	if min == 0 { min = tls.VersionTLS13 }
	clientAuth := tls.RequireAndVerifyClientCert
	if !c.RequireClient { clientAuth = tls.VerifyClientCertIfGiven }
	return &tls.Config{MinVersion:min, Certificates:[]tls.Certificate{cert}, ClientCAs:pool, ClientAuth:clientAuth}, nil
}

func (c MTLSConfig) ClientTLS() (*tls.Config, error) {
	cert, pool, err := LoadMTLSConfig(c)
	if err != nil { return nil, err }
	min := c.MinVersion
	if min == 0 { min = tls.VersionTLS13 }
	return &tls.Config{MinVersion:min, Certificates:[]tls.Certificate{cert}, RootCAs:pool, ServerName:c.ServerName}, nil
}
