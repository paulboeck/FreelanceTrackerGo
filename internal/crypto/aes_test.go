package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := EncryptionKey("test-seed")
	plaintext := "my-secret-password"

	// Test encryption
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if encrypted == plaintext {
		t.Error("encrypted text should be different from plaintext")
	}

	// Test decryption
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("decrypted text doesn't match original. Got %s, want %s", decrypted, plaintext)
	}
}

func TestEncryptionKeyConsistency(t *testing.T) {
	seed := "consistent-seed"
	key1 := EncryptionKey(seed)
	key2 := EncryptionKey(seed)

	if len(key1) != 32 {
		t.Errorf("key length should be 32 bytes, got %d", len(key1))
	}

	for i, b := range key1 {
		if key2[i] != b {
			t.Error("keys generated from same seed should be identical")
			break
		}
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := EncryptionKey("seed1")
	key2 := EncryptionKey("seed2")
	plaintext := "secret"

	encrypted, err := Encrypt(plaintext, key1)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Error("decryption should fail with wrong key")
	}
}