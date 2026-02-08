package pki

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

type CertSpec struct {
	CommonName string
	Orgs       []string

	DNSNames []string
	IPs      []net.IP

	IsCA       bool
	ValidYears int

	KeyUsage    x509.KeyUsage
	ExtKeyUsage []x509.ExtKeyUsage
}

// GenerateCA 生成CA证书
func GenerateCA(spec CertSpec) (*x509.Certificate, *rsa.PrivateKey, []byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:   spec.CommonName,
			Organization: spec.Orgs,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(spec.ValidYears, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	cert, _ := x509.ParseCertificate(der)
	return cert, key, certPEM, keyPEM, nil
}

// IssueCert 用这个生成证书
/*
	subject:
	kube-apiserver:
		ExtKeyUsage = serverAuth
		SAN 包含：
		所有 master IP
		LB / VIP
		kubernetes
		kubernetes.default
		kubernetes.default.svc
		Service IP（如 10.96.0.1）
	kubelet:
		subject:
			CN = system:node:<nodeName>
			O = system:nodes
	controller/scheduler:
		subject:
			CN = system:kube-controller-manager
			CN = system:kube-scheduler
	kubectl:
		CN = admin
		O  = system:masters
*/
func IssueCert(
	caCert *x509.Certificate,
	caKey *rsa.PrivateKey,
	spec CertSpec,
) ([]byte, []byte, error) {

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName:   spec.CommonName,
			Organization: spec.Orgs,
		},
		DNSNames:    spec.DNSNames,
		IPAddresses: spec.IPs,

		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(spec.ValidYears, 0, 0),

		KeyUsage:    spec.KeyUsage,
		ExtKeyUsage: spec.ExtKeyUsage,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return certPEM, keyPEM, nil
}

// WriteFile writes data to file with 0600 permissions
func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0600)
}

// ParseCert parses a PEM encoded certificate
func ParseCert(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}
	return x509.ParseCertificate(block.Bytes)
}

// ParseRSAKey parses a PEM encoded RSA private key
func ParseRSAKey(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to parse key PEM")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
