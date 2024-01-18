package utils

import (
	"testing"
)

func TestWebhookHandler(t *testing.T) {
	t.Run("Encrypts password correctly", func(t *testing.T) {
		key := []byte("byteslongpassphraseforencryption") // Key should be 16, 24, or 32 bytes long
		password := "mySuperSecretPassword"
		encryptedPassword, err := HashString(key, password)
		if err != nil {
			t.Errorf("Error encrypting password %s", err)
		}
		if encryptedPassword == password {
			t.Errorf("Password was not encrypted %s", encryptedPassword)
		}
	})

	t.Run("Encrypts then decrypts password correctly", func(t *testing.T) {
		key := []byte("byteslongpassphraseforencryption") // Key should be 16, 24, or 32 bytes long
		password := "mySuperSecretPassword"
		encryptedPassword, err := HashString(key, password)
		if err != nil {
			t.Errorf("Error encrypting password %s", err)
		}
		if encryptedPassword == password {
			t.Errorf("Password was not encrypted %s", encryptedPassword)
		}
		decryptedPassword, err := RestoreHashedString(key, encryptedPassword)
		if err != nil {
			t.Errorf("Error decrypting password %s", err)
		}
		if decryptedPassword != password {
			t.Errorf("Password was not decrypted " + encryptedPassword)
		}
	})
}
