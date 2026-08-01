package release

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

func SignFile(path, privateKeyPath string) (string, error) {
	info, err := os.Stat(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("inspect signing key: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return "", fmt.Errorf("signing key permissions must not allow group or other access")
	}
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("read signing key: %w", err)
	}
	block, rest := pem.Decode(keyData)
	if block == nil || len(strings.TrimSpace(string(rest))) != 0 {
		return "", fmt.Errorf("signing key must contain one PEM PKCS#8 key")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse signing key: %w", err)
	}
	privateKey, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return "", fmt.Errorf("signing key must be Ed25519")
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read signed file: %w", err)
	}
	signature := ed25519.Sign(privateKey, payload)
	return base64.StdEncoding.EncodeToString(signature), nil
}

func VerifyFile(path, signature string, publicKey ed25519.PublicKey) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read signed file: %w", err)
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(strings.TrimSpace(signature))
	if err != nil {
		return fmt.Errorf("decode signature: %w", err)
	}
	if !ed25519.Verify(publicKey, payload, decoded) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

func LoadPublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read verification key: %w", err)
	}
	block, rest := pem.Decode(data)
	if block == nil || len(strings.TrimSpace(string(rest))) != 0 {
		return nil, fmt.Errorf("verification key must contain one PEM public key")
	}
	parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse verification key: %w", err)
	}
	publicKey, ok := parsed.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("verification key must be Ed25519")
	}
	return publicKey, nil
}
