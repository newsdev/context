package crypter

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/nytinteractive/context/crypter/std"
)

func randomStdCrypter() (Crypter, error) {
	key := make([]byte, std.SymetricKeyLength+std.HmacKeyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return std.New(key[:std.SymetricKeyLength], key[std.SymetricKeyLength:])
}

func TestStdCrypterEncodeDecode(t *testing.T) {
	c, err := randomStdCrypter()
	if err != nil {
		t.Fatal(err)
	}

	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[]✓.")

	cipherbytes, err := c.EncryptAndSign(originalbytes)
	if err != nil {
		t.Error(err)
	}

	if bytes.Contains(cipherbytes, originalbytes) {
		t.Error("encoding the bytes didn't work!")
	}

	plainbytes, err := c.ValidateAndDecrypt(cipherbytes)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(plainbytes, originalbytes) {
		t.Errorf("decoded bytes did not match!", originalbytes, plainbytes)
	}
}
