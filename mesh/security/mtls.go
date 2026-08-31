package security

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type MTLSConfig struct {
	CertFile string
	KeyFile string
	CAFile string
	MinVersion uint16
	ServerName string
	RequireClient bool
	SPIFFETrustDomain string
}

func (c MTLSConfig) Validate() error {
	if strings.TrimSpace(c.CertFile)==""||strings.TrimSpace(c.KeyFile)==""||strings.TrimSpace(c.CAFile)==""{return errors.New("mTLS certificate, key and CA files are required")}
	if c.MinVersion!=0&&c.MinVersion!=tls.VersionTLS13{return errors.New("TLS 1.3 is required")}
	if d:=strings.TrimSpace(c.SPIFFETrustDomain);d!=""&&!validSPIFFETrustDomain(d){return fmt.Errorf("invalid SPIFFE trust domain %q",d)}
	return nil
}

func LoadMTLSConfig(c MTLSConfig)(tls.Certificate,*x509.CertPool,error){if err:=c.Validate();err!=nil{return tls.Certificate{},nil,err};cert,err:=tls.LoadX509KeyPair(c.CertFile,c.KeyFile);if err!=nil{return tls.Certificate{},nil,err};caPEM,err:=os.ReadFile(c.CAFile);if err!=nil{return tls.Certificate{},nil,err};pool:=x509.NewCertPool();if !pool.AppendCertsFromPEM(caPEM){return tls.Certificate{},nil,errors.New("CA file contains no valid certificates")};return cert,pool,nil}

func (c MTLSConfig) ServerTLS()(*tls.Config,error){cert,pool,err:=LoadMTLSConfig(c);if err!=nil{return nil,err};clientAuth:=tls.RequireAndVerifyClientCert;if !c.RequireClient{clientAuth=tls.VerifyClientCertIfGiven};cfg:=&tls.Config{MinVersion:tls.VersionTLS13,Certificates:[]tls.Certificate{cert},ClientCAs:pool,ClientAuth:clientAuth};if strings.TrimSpace(c.SPIFFETrustDomain)!=""{cfg.VerifyConnection=spiffeVerifier(pool,c.SPIFFETrustDomain)};return cfg,nil}
func (c MTLSConfig) ClientTLS()(*tls.Config,error){cert,pool,err:=LoadMTLSConfig(c);if err!=nil{return nil,err};cfg:=&tls.Config{MinVersion:tls.VersionTLS13,Certificates:[]tls.Certificate{cert},RootCAs:pool,ServerName:c.ServerName};if strings.TrimSpace(c.SPIFFETrustDomain)!=""{cfg.VerifyConnection=spiffeVerifier(pool,c.SPIFFETrustDomain)};return cfg,nil}

func spiffeVerifier(pool *x509.CertPool,trustDomain string)func(tls.ConnectionState)error{return func(cs tls.ConnectionState)error{if len(cs.PeerCertificates)==0{return errors.New("peer certificate missing")};cert:=cs.PeerCertificates[0];opts:=x509.VerifyOptions{Roots:pool,Intermediates:x509.NewCertPool(),DNSName:"",KeyUsages:[]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth,x509.ExtKeyUsageServerAuth}};for _,c:=range cs.PeerCertificates[1:]{opts.Intermediates.AddCert(c)};if _,err:=cert.Verify(opts);err!=nil{return fmt.Errorf("peer certificate verification failed: %w",err)};prefix:="spiffe://"+strings.TrimSuffix(trustDomain,"/")+"/";for _,u:=range cert.URIs{if strings.HasPrefix(strings.ToLower(u.String()),strings.ToLower(prefix)){return nil}};return fmt.Errorf("peer certificate lacks SPIFFE identity in trust domain %q",trustDomain)}}

func validSPIFFETrustDomain(s string)bool{u,err:=url.Parse("spiffe://"+strings.TrimPrefix(s,"spiffe://"));return err==nil&&u.Host!=""&&!strings.ContainsAny(u.Host," /:@")}
