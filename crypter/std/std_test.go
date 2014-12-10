package std

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func randomStdCrypter() (*stdCrypter, error) {
	key := make([]byte, SymetricKeyLength+HmacKeyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return New(key[:SymetricKeyLength], key[SymetricKeyLength:])
}

func TestEncodeDecode(t *testing.T) {

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

// func TestSeperateEncoding(t *testing.T) {

// 	key := make([]byte, HmacKeyLength+SymetricKeyLength)
// 	if _, err := io.ReadFull(rand.Reader, key); err != nil {
// 		t.Fatal(err)
// 	}

// 	c1, err := NewstdCrypter(bytes.NewBuffer(key))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	c2, err := NewstdCrypter(bytes.NewBuffer(key))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")
// 	fmt.Println(originalbytes)

// 	cipherbytes1, err := c1.EncryptAndSign(originalbytes)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	fmt.Println(cipherbytes1)

// 	cipherbytes2, err := c2.EncryptAndSign(originalbytes)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	fmt.Println(cipherbytes2)

// 	if bytes.Equal(cipherbytes1, cipherbytes2) {
// 		t.Error("seperate encodings of the same string matched!")
// 	}
// }

// func TestRepeatedEncoding(t *testing.T) {

// 	c, err := NewRandomstdCrypter()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")
// 	fmt.Println(originalbytes)

// 	cipherbytes1, err := c.EncryptAndSign(originalbytes)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	fmt.Println(cipherbytes1)

// 	cipherbytes2, err := c.EncryptAndSign(originalbytes)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	fmt.Println(cipherbytes2)

// 	if bytes.Equal(cipherbytes1, cipherbytes2) {
// 		t.Error("repeated encodings of the same string matched!")
// 	}
// }

// func BenchmarkEncode(b *testing.B) {
// 	b.StopTimer()

// 	c, err := NewRandomstdCrypter()
// 	if err != nil {
// 		b.Fatal(err)
// 	}

// 	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")

// 	for i := 0; i < b.N; i++ {
// 		b.StartTimer()
// 		c.EncryptAndSign(originalbytes)
// 		b.StopTimer()
// 	}
// }

// func BenchmarkEncodeCold(b *testing.B) {
// 	b.StopTimer()

// 	originalbytes := []byte("Test message !@#$%^&*()_1234567890{}[].")

// 	for i := 0; i < b.N; i++ {
// 		c, err := NewRandomstdCrypter()
// 		if err != nil {
// 			b.Fatal(err)
// 		}
// 		b.StartTimer()
// 		c.EncryptAndSign(originalbytes)
// 		b.StopTimer()
// 	}
// }
