package crypter

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/buth/context/crypter/std"
)

type Crypter interface {
	EncryptAndSign([]byte) ([]byte, error)
	ValidateAndDecrypt([]byte) ([]byte, error)
}

func NewCrypter(kind string, key []byte) (Crypter, error) {

	// Select a crypter based on kind.
	switch kind {
	case "std":
		return std.New(key[:std.SymetricKeyLength], key[std.SymetricKeyLength:])
	}

	// Assuming no crypter is implemented for kind.
	return nil, NoCrypterError{kind}
}

func NewKey(kind string) ([]byte, error) {

	// Select a crypter based on kind.
	switch kind {
	case "std":
		key := make([]byte, std.SymetricKeyLength+std.HmacKeyLength)
		if _, err := io.ReadFull(rand.Reader, key); err != nil {
			return nil, err
		}
		return key, nil
	}

	// Assuming no crypter is implemented for kind.
	return nil, NoCrypterError{kind}
}

type NoCrypterError struct {
	Kind string
}

func (e NoCrypterError) Error() string {
	return fmt.Sprintf("crypter: crypter \"%s\" has not been implemented", e.Kind)
}
