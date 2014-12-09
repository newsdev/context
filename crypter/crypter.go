package crypter

type Crypter interface {
	EncryptAndSign([]byte) ([]byte, error)
	ValidateAndDecrypt([]byte) ([]byte, error)
}
