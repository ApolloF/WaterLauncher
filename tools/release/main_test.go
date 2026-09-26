package main

import (
	"bytes"
	"crypto/ed25519"
	"testing"
)

func TestBackupRoundTrip(t *testing.T) {
	_, k, _ := ed25519.GenerateKey(nil)
	raw, err := sealBackup(k, "correct horse battery staple", 100_000)
	if err != nil {
		t.Fatal(err)
	}
	seed, err := openBackup(raw, "correct horse battery staple")
	if err != nil || !bytes.Equal(seed, k.Seed()) {
		t.Fatalf("restore: %v", err)
	}
	if _, err := openBackup(raw, "wrong password here"); err == nil {
		t.Error("a wrong password opened the backup")
	}
	bad := bytes.Replace(raw, []byte(`"box": "`), []byte(`"box": "AA`), 1)
	if _, err := openBackup(bad, "correct horse battery staple"); err == nil {
		t.Error("a damaged backup opened")
	}
}
