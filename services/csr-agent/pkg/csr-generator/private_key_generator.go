package csr_generator

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/rotisserie/eris"
)

func NewPrivateKeyGenerator() PrivateKeyGenerator {
	return &privateKeyGenerator{}
}

type privateKeyGenerator struct {
}

func (p *privateKeyGenerator) GenerateRSA(keySize int) ([]byte, error) {
	priv, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, eris.Errorf("RSA key generation failed (%v)", err)
	}
	privKey := x509.MarshalPKCS1PrivateKey(priv)
	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKey,
	}
	return pem.EncodeToMemory(keyBlock), nil
}
