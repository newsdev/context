package std

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"io"
)

const (

	// SymetricKeyLength is the length in bytes of the key used with the
	// AES-256 algorithm
	SymetricKeyLength = 32

	// HmacKeyLength is the length in bytes of the key used in the HMAC
	// SHA-512 algorithm
	HmacKeyLength = 128
)

// A stdCrypter is an enstdCrypter/destdCrypter set to use a specific encryption key (for
// AES-256 in CBC mode) and signing key (for HMAC SHA-512) combination.
type stdCrypter struct {
	block   cipher.Block
	hmacKey []byte
}

func New(cipherKey, hmacKey []byte) (*stdCrypter, error) {

	// Confirm that there are enough bytes in the cipher key to select AES-256
	// and no more.
	if len(cipherKey) != SymetricKeyLength {
		return nil, stdCrypterError{"cipher key has the wrong length for AES-256 (32 bytes)"}
	}

	// Create the block from the cipher key. An important assumption we are
	// making is that the resulting block does not contain references to the
	// original, mutable cipher key.
	b, err := aes.NewCipher(cipherKey)
	if err != nil {
		return nil, err
	}

	// Confirm that there are enough bytes in the HMAC key to use it with the
	// SHA-512 algorithm without resorting to padding.
	if len(hmacKey) != HmacKeyLength {
		return nil, stdCrypterError{"HMAC key has the wrong length for SHA-512 (128 bytes)"}
	}

	// Copy the HMAC key to insure immutability of the stdCrypter.
	h := make([]byte, 128)
	copy(h, hmacKey)

	return &stdCrypter{
		block:   b,
		hmacKey: h,
	}, nil
}

func NewRandom() ([]byte, []byte, *stdCrypter, error) {

	symetricKey := make([]byte, SymetricKeyLength)
	if _, err := io.ReadFull(rand.Reader, symetricKey); err != nil {
		return nil, nil, nil, err
	}

	hmacKey := make([]byte, HmacKeyLength)
	if _, err := io.ReadFull(rand.Reader, hmacKey); err != nil {
		return nil, nil, nil, err
	}

	c, err := New(symetricKey, hmacKey)
	if err != nil {
		return nil, nil, nil, err
	}

	return symetricKey, hmacKey, c, nil
}

// hmac computes and returns SHA-512 HMAC sum using the signing key.
func (c *stdCrypter) hmac(message []byte) []byte {
	signer := hmac.New(sha512.New, c.hmacKey)
	signer.Write(message)
	return signer.Sum(nil)
}

// encrypt encrypts a slice of bytes using the AES-256 cipher in CBC mode and
// returns an usigned sice of cipher bytes that begins with the IV.
func (c *stdCrypter) encrypt(plainbytes []byte) ([]byte, error) {

	// Initialize size with room for the IV.
	size := aes.BlockSize + len(plainbytes)

	// Add extra padding if the size is not a multiple of the Block size.
	if extra := len(plainbytes) % aes.BlockSize; extra != 0 {
		size += aes.BlockSize - extra
	}

	// Create the cipherbytes slice and copy in the plainbytes.
	cipherbytes := make([]byte, size)
	copy(cipherbytes[aes.BlockSize:], plainbytes)

	// Use an IV at the front of the cipherbytes, and attempt to read in random bits.
	iv := cipherbytes[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return []byte{}, err
	}

	// Create the enstdCrypter and crypt the plainbytes in place.
	mode := cipher.NewCBCEncrypter(c.block, iv)
	mode.CryptBlocks(cipherbytes[aes.BlockSize:], cipherbytes[aes.BlockSize:])

	return cipherbytes, nil
}

// decrypt decrypts a slice of cipherbytes using the AES-256 cipher in CBC
// mode and returns a slice of plain bytes. The first Block of the cipherbytes
// argument is expected to be the IV. It does not verify or expect a signature
// to be present in the cipherbytes argument.
func (c *stdCrypter) decrypt(cipherbytes []byte) ([]byte, error) {

	// We need an IV and at least one Block of cipherbytes to proceed.
	if len(cipherbytes) < aes.BlockSize*2 {
		return []byte{}, stdCrypterError{"cipherbytes is too short"}
	}

	// CBC mode always works in whole Blocks.
	if len(cipherbytes)%aes.BlockSize != 0 {
		return []byte{}, stdCrypterError{"cipherbytes is not a multiple of the block size"}
	}

	// IV is the first BlockSize bytes of the message.
	iv := cipherbytes[:aes.BlockSize]

	// Allocate a new byte array to hold the plainbytes
	plainbytes := make([]byte, len(cipherbytes)-aes.BlockSize)

	// Decrypt the cipherbytes and trim the result.
	mode := cipher.NewCBCDecrypter(c.block, iv)
	mode.CryptBlocks(plainbytes, cipherbytes[aes.BlockSize:])
	plainbytes = bytes.TrimRight(plainbytes, "\x00")

	return plainbytes, nil
}

// EncryptAndSign converts plainbytes to signed cipherbytes by encrypting the
// plainbytes using AES-256 and prepending a Hmac SHA-512 signature.
func (c *stdCrypter) EncryptAndSign(plainbytes []byte) ([]byte, error) {

	// Encrypt the slice of plainbytes, producing cipherbytes.
	cipherbytes, err := c.encrypt(plainbytes)
	if err != nil {
		return nil, err
	}

	// Get the signatrue for the cipherbytes.
	hmacbytes := c.hmac(cipherbytes)

	// Copy all the bytes into a single byte string.
	messagebytes := make([]byte, sha512.Size+len(cipherbytes))
	copy(messagebytes[:len(hmacbytes)], hmacbytes)
	copy(messagebytes[len(hmacbytes):], cipherbytes)

	return messagebytes, nil
}

// Decrypt converts signed slice of cipherbytes to plainbytes by first
// validating a prepended Hmac SHA-512 signature and then decrypting the
// remaining message using AES-256.
func (c *stdCrypter) ValidateAndDecrypt(messagebytes []byte) ([]byte, error) {

	// Check that message bytes is long enough.
	if len(messagebytes) < 64 {
		return nil, stdCrypterError{"message signature is too short"}
	}

	// Check the signature.
	if hmac.Equal(messagebytes[:64], c.hmac(messagebytes[64:])) != true {
		return nil, stdCrypterError{"invalid signature"}
	}

	// Decode the encrypted bytes.
	plainbytes, err := c.decrypt(messagebytes[64:])
	if err != nil {
		return nil, err
	}

	return plainbytes, nil
}

// stdCrypterError represents a run-time error in a stdCrypter method.
type stdCrypterError struct {
	Err string
}

func (e stdCrypterError) Error() string {
	return fmt.Sprintf("stdCrypter: %s", e.Err)
}
