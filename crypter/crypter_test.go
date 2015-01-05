package crypter

import (
	"bytes"
	"testing"
)

var kinds = []string{"std"}
var message = []byte("Test message !@#$%^&*()_1234567890{}[]✓.")

func TestCrypterKeyGeneration(t *testing.T) {
	for _, kind := range kinds {

		k1, err := NewKey(kind)
		if err != nil {
			t.Error(err)
			continue
		}

		if _, err := NewCrypter(kind, k1); err != nil {
			t.Error(err)
			continue
		}

		k2, err := NewKey(kind)
		if err != nil {
			t.Error(err)
			continue
		}

		if bytes.Equal(k1, k2) {
			t.Error("identical keys generated!")
			continue
		}

		if _, err := NewCrypter(kind, k1); err != nil {
			t.Error(err)
		}
	}
}

func TestCrypterEncodeDecode(t *testing.T) {
	for _, kind := range kinds {

		k, err := NewKey(kind)
		if err != nil {
			t.Error(err)
			continue
		}

		c, err := NewCrypter(kind, k)
		if err != nil {
			t.Error(err)
			continue
		}

		cipherbytes, err := c.EncryptAndSign(message)
		if err != nil {
			t.Error(err)
			continue
		}

		if bytes.Contains(cipherbytes, message) {
			t.Error("encoding the bytes didn't work!")
			continue
		}

		plainbytes, err := c.ValidateAndDecrypt(cipherbytes)
		if err != nil {
			t.Error(err)
			continue
		}

		if !bytes.Equal(plainbytes, message) {
			t.Errorf("decoded bytes did not match!", message, plainbytes)
		}
	}
}

func TestCrypterMultipleEncodings(t *testing.T) {
	for _, kind := range kinds {

		k, err := NewKey(kind)
		if err != nil {
			t.Error(err)
			continue
		}

		c, err := NewCrypter(kind, k)
		if err != nil {
			t.Error(err)
			continue
		}

		cipherbytes1, err := c.EncryptAndSign(message)
		if err != nil {
			t.Error(err)
			continue
		}

		cipherbytes2, err := c.EncryptAndSign(message)
		if err != nil {
			t.Error(err)
			continue
		}

		if bytes.Equal(cipherbytes1, cipherbytes2) {
			t.Errorf("sequential encodings returned the same result!", cipherbytes1, cipherbytes2)
		}
	}
}
