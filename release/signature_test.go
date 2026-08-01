package release

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestSignAndVerifyFile(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	keyPath := filepath.Join(directory, "signing-key.pem")
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}), 0o600); err != nil {
		t.Fatal(err)
	}
	payloadPath := filepath.Join(directory, "provenance.json")
	if err := os.WriteFile(payloadPath, []byte("provenance"), 0o600); err != nil {
		t.Fatal(err)
	}
	signature, err := SignFile(payloadPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyFile(payloadPath, signature, publicKey); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(payloadPath, []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFile(payloadPath, signature, publicKey); err == nil {
		t.Fatal("tampered provenance passed verification")
	}
}

func TestSignRejectsLoosePrivateKeyPermissions(t *testing.T) {
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	encoded, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	directory := t.TempDir()
	keyPath := filepath.Join(directory, "key.pem")
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: encoded}), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(directory, "payload")
	_ = os.WriteFile(payload, []byte("data"), 0o600)
	if _, err := SignFile(payload, keyPath); err == nil {
		t.Fatal("loosely permissioned signing key accepted")
	}
}
