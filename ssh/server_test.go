package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"testing"
)

var clientPublicKey = []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAAgQCvU24CoYxhz0r54iXINU3S2v/ZeqgQminFfyLjdWb7ed++ShkFQ08DEgTwXLc1UepvUsPuXYbNyG0d3j8Ib+DWDVaXdqAEH1v5gL7Ql0zuI8SYJ/ybTsam4Uzw+beWO4oyAkAAdVBMKm/obDOEk85Px8R6u28fowefmOic3kFLSw== test`)

var clientPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCvU24CoYxhz0r54iXINU3S2v/ZeqgQminFfyLjdWb7ed++ShkF
Q08DEgTwXLc1UepvUsPuXYbNyG0d3j8Ib+DWDVaXdqAEH1v5gL7Ql0zuI8SYJ/yb
Tsam4Uzw+beWO4oyAkAAdVBMKm/obDOEk85Px8R6u28fowefmOic3kFLSwIDAQAB
AoGBAIU75m7Ta0Xs7HImnEWf1Es3J5SSdGNhc/rkmZO25RKX1CLcVlU8iC+yItSx
8HvxizEb+U8L/eQlul4nRUlZE8b0dBKv20gAn2cOXaCQUCxEzTlk9g9N6Iwg4XIl
S77ifXFpsTWlaw2yvrF4uNEAI7MJioMd7TJIA9s6hWvmrSjRAkEA2V8o/DgJKryg
RmSqVAG3gaH10nT2nOugkf59UjIwiC2DqNdCJ2LWYui6il0nSnKsVFghsgiIFApF
/9kwpmTmuQJBAM57fTM/9DEEsgbo/gfJVOeIA6nw1QS3mly+yWZ3c2nitWASpxZ5
4DGm1PAxTqMWdxbZf2gNP4vKrRbNiwdDwCMCQEU6UUtCbWj2+fRxSu3GPjNC6Y9F
QOVpBZJ5gmATK/GyzSOQqrjweWa2x/IZCNJlAw05pEGXBf+b5f89pIjZycECQCUP
jypKuVavBBEvcqENJvsjs5ymCGX/Wmp5KAcHO6TutyVWU706BN6ElkXCY93r41Yr
la2kaxp5N1YXcHPOWkcCQHerVx2JJcrcvKEURWsOQbgCMa0U0UzjC7QYGIAmCxTJ
48uysNydUBwErvU2ZiVYJ9Vq7bvQonfr8nUYyfEyO0g=
-----END RSA PRIVATE KEY-----
`)

var serverHostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDcyLyJpKhqmaewI2uxKB8DFx6bwT+AmJAt3a2eP0LX32C7EPfS
EksfvJIrY3Qyl5xkeT/jVvWV5UCIddsAo7zAU070cKEBwWNobLfByYTk183g4hzk
eFCoXYqxRa6fUEHNIitgCSh5IumtFYm/+VSYvIEagDsrEFi3AoCfs3/poQIDAQAB
AoGAeLPRx1parLTGZVhoBtlNYw4fsL1Mr0w4/qpDwdEKBSTdKEgVhCZ2JaqdKtVb
cFFMse1LzUj8SG+mATSVf1sE1AhXaetkGwOaTZ7g8EZ5GAMUdQ9Rt3A6BlcgqXwL
HXanMULNCO17Z0AsoPaoImN0vbte01Ig3roBeSl8S9l3lcECQQD2qMj0H6tY+391
4FkKT5poPAbUpCxigcAgX+sVVAFEhkWCv/XUaGajSUS9fIoqzCMnDxzrcv882geo
cGdcVfnZAkEA5SUbYf81f/QmxlY7kq11u8wC8C2u+cSK9Y8zXDaRJL4fwVPwLAuV
Q1vkiYQeJwYIP9nJ9WGyqx3jyFC6982JCQJBALTST18HyGlXFb2oVh4E9UDsoGVK
ZW9hhyM0rfXYu4UsmdCcQO8SCgwyLj5rCi8Nr8d2gNDqYMqPW4XTwTIjpSECQQCF
xs0ewDUGt45vmmZ7MoOamPdaKwGNVf5ecDTm8AB6t/ioEI4V2MlSovJgil5kH/Ru
+oIanOgHWJLkHqWZCEipAkA4R01gRsK1cFq+g0IyOTT3yHMf8m9hZlLiiKMnzEKw
ho2oAmqAItSDs7nmeavM5YL/jtzB2izZp2XsnYdb5tsN
-----END RSA PRIVATE KEY-----
`)

type handler struct{}

func (h handler) ServeSSH(stdout io.Writer, env map[string]string, command string) error {
	fmt.Println("HELLO", command)
	return nil
}

func newTestServer() (*Server, error) {

	hostKey, err := ssh.ParsePrivateKey(serverHostKey)
	if err != nil {
		return nil, err
	}

	clientKey, _, _, _, err := ssh.ParseAuthorizedKey(clientPublicKey)
	if err != nil {
		return nil, err
	}

	return &Server{
		Addr:    `:2222`,
		Handler: handler{},
		HostKey: hostKey,
		Commands: func(addr net.Addr, key ssh.PublicKey) ([]string, error) {
			if bytes.Equal(key.Marshal(), clientKey.Marshal()) {
				return []string{"test"}, nil
			}
			return nil, errors.New("unauthorized")
		},
	}, nil
}

func TestServer(t *testing.T) {

	server, err := newTestServer()
	if err != nil {
		t.Fatal(err)
	}

	go server.ListenAndServe()

	config := &ssh.ClientConfig{
		User: `context`,
	}

	privateKeyParsed, err := ssh.ParsePrivateKey(clientPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	config.Auth = []ssh.AuthMethod{
		ssh.PublicKeys(privateKeyParsed),
	}

	client, err := ssh.Dial("tcp", `:2222`, config)
	if err != nil {
		t.Fatal(err)
	}

	session, err := client.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	if err := session.Run(`test A=B`); err != nil {
		t.Fatal(err)
	}
}
