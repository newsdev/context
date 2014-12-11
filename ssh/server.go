package ssh

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
)

const (
	User = `context`
)

type Handler interface {
	ServeSSH(stdout io.Writer, env map[string]string, command string) error
}

type Server struct {
	Addr     string
	Handler  Handler
	HostKey  ssh.Signer
	Commands map[string][][]byte
}

// ListenAndServe starts a new SSH server listening on the given address.
func (s *Server) ListenAndServe() error {

	certChecker := &ssh.CertChecker{
		IsAuthority: func(k ssh.PublicKey) bool { return false },
		UserKeyFallback: func(conn ssh.ConnMetadata, userPublicKey ssh.PublicKey) (*ssh.Permissions, error) {
			if conn.User() != User {
				return nil, ServerError{"unauthorized"}
			}
			userPublicKeyBytes := userPublicKey.Marshal()
			permissions := &ssh.Permissions{Extensions: make(map[string]string)}
			for command, publicKeys := range s.Commands {
				for _, publicKey := range publicKeys {
					if bytes.Equal(userPublicKeyBytes, publicKey) {
						permissions.Extensions[command] = ""
						break
					}
				}
			}
			if len(permissions.Extensions) == 0 {
				return nil, ServerError{"unauthorized"}
			}
			return permissions, nil
		},
	}

	serverConfig := &ssh.ServerConfig{
		Config: ssh.Config{
			Ciphers: []string{"aes256-ctr"},
			MACs:    []string{"hmac-sha1"},
		},
		PublicKeyCallback: certChecker.Authenticate,
	}
	serverConfig.AddHostKey(s.HostKey)

	// Start listening.
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	for {

		// Accept a new connection.
		nConn, err := listener.Accept()
		if err != nil {
			return err
		}

		sConn, channels, reqs, err := ssh.NewServerConn(nConn, serverConfig)
		if err != nil {
			log.Println("server: failed to handshake")
			continue
		}

		// The incoming Request channel must be serviced.
		go ssh.DiscardRequests(reqs)

		// Iterate through the incomming channel requests.
		for newChannel := range channels {

			// Only accept sessions.
			if newChannel.ChannelType() != "session" {
				newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
				continue
			}

			// Attempt to accecpt the session channel.
			channel, requests, err := newChannel.Accept()
			if err != nil {
				log.Printf("server: could not accept channel: %s\n", err.Error())
				continue
			}

			// Create an environment specific to the channel.
			env := make(map[string]string)

			// Iterate through the requests on the channel.
			for request := range requests {

				// Unpack the payload.
				payload, err := UnpackMessage(request.Payload)
				if err != nil {
					log.Println(err)
				}

				// Switch on the request type.
				switch request.Type {
				case "env":
					for i := 0; i < len(payload)/2; i++ {
						env[payload[i*2]] = payload[i*2+1]
					}

				case "exec":

					// Indicate that we have started running the command.
					request.Reply(true, nil)

					// The exit status will be reported as a 4-byte, little-endian integer.
					exitStatusBuffer := bytes.NewBuffer([]byte{})

					// Run the command, reporting any error as a failure.
					if err := s.Handler.ServeSSH(channel, env, payload[0]); err != nil {
						log.Println(err)
						binary.Write(exitStatusBuffer, binary.BigEndian, uint32(1))
					} else {
						binary.Write(exitStatusBuffer, binary.BigEndian, uint32(0))
					}

					// Write the exit status.
					channel.SendRequest("exit-status", false, exitStatusBuffer.Bytes())

					// Only one exec command can be handled per channel, so we're done.
					break
				}
			}
		}

		// Close the connection after we've handled all channels.
		sConn.Close()
	}

	return nil
}

// // AddWriteKey adds a public key that is authorized to connect to the server
// // and perform both read and write operations. If the has already been added
// // to the server, this function will update its status.
// func (s *Server) AddWriteKey(key ssh.PublicKey) {
// 	s.writeKeysMu.Lock()
// 	defer s.writeKeysMu.Unlock()
// 	s.writeKeys[string(key.Marshal())] = true
// }

// func (s *Server) matchWriteKey(key ssh.PublicKey) bool {
// 	s.writeKeysMu.RLock()
// 	defer s.writeKeysMu.RUnlock()
// 	return s.writeKeys[string(key.Marshal())]
// }

// // AddReaderKey adds a public key that is authorized to connect to the server
// // and perform only read operations. If the has already been added to the
// // server, this function will update its status. to the server, this function
// // will update its status.
// func (s *Server) AddReadKey(key ssh.PublicKey) {
// 	s.readKeysMu.Lock()
// 	s.readKeys.PushBack(key.Marshal())

// 	s.readKeysMu.Unlock()
// }

// func (s *Server) matchReadKey(key ssh.PublicKey) bool {
// 	s.readKeysMu.RLock()
// 	defer s.readKeysMu.RUnlock()
// 	return matchKey(key, s.readKeys)
// }

// // checkUserKey determines whether or not the given public key is present for
// // the user indicated in the SSH connection meta-data.
// func (s *Server) checkUserKey(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {

// 	// Writers must match one of the write keys.
// 	if conn.User() == WriterUser {
// 		if s.matchWriteKey(key) {
// 			return &ssh.Permissions{}, nil
// 		}
// 		return nil, ServerError{"unauthorized"}
// 	}

// 	// Readers can match the registration key.
// 	if conn.User() == ReadUser && s.matchReadKey(key) {
// 		return &ssh.Permissions{}, nil
// 	}

// 	// Parse out the host.
// 	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Get the encrypted public key from the backend.
// 	pubKeyEncrypted, err := s.backend.GetVariable(SSHGroupName, host)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Decrypt the public key.
// 	pubKey, err := s.crypter.ValidateAndDecrypt(pubKeyEncrypted)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Check if it matches.
// 	if bytes.Equal(pubKey, key.Marshal()) {
// 		return &ssh.Permissions{}, nil
// 	}

// 	// The default case is to return an error.
// 	return nil, ServerError{"unauthorized"}
// }

// func (s *Server) exec(stdout io.Writer, canWrite bool, raddr string, environment map[string]string, commandString string) error {

// 	// Try to pull the group from the environment.
// 	var group string
// 	if environmentGroup, ok := environment["GROUP"]; ok {
// 		group = environmentGroup
// 	}

// 	// The first (and only) payload value should be a string.
// 	components := strings.SplitN(commandString, ` `, 2)
// 	command := components[0]
// 	argument := ""
// 	if len(components) == 2 {
// 		argument = components[1]
// 	}

// 	switch command {

// 	case "env":

// 		// Pull the encrypted values from the store.
// 		variables, err := s.backend.GetGroup(group)
// 		if err != nil {
// 			return err
// 		}

// 		for variable, cryptedValue := range variables {

// 			// Attempt to decrypt the encrypted value.
// 			value, err := s.crypter.ValidateAndDecrypt(cryptedValue)
// 			if err != nil {
// 				return err
// 			}

// 			// Write the variable to the channel.
// 			fmt.Fprintf(stdout, "%s=%s\n", variable, value)
// 		}

// 	case "export":

// 		// Check for write permission.
// 		if !canWrite {
// 			return ServerError{"unauthorized"}
// 		}

// 		// Parse the variable name and value from the argument.
// 		argumentComponents := strings.SplitN(argument, `=`, 2)
// 		variable := argumentComponents[0]
// 		value := ""

// 		// We may need to check the environment for the value.
// 		if len(argumentComponents) == 2 {
// 			value = argumentComponents[1]
// 		} else if environmentValue, ok := environment[variable]; ok {
// 			value = environmentValue
// 		}

// 		// Attempt to encrypt the value.
// 		cryptedValue, err := s.crypter.EncryptAndSign([]byte(value))
// 		if err != nil {
// 			return err
// 		}

// 		// Save the encrypted value in the store.
// 		if err := s.backend.SetVariable(group, variable, cryptedValue); err != nil {
// 			return err
// 		}

// 	case "unset":

// 		// Check for write permission.
// 		if !canWrite {
// 			return ServerError{"unauthorized"}
// 		}

// 		// Assume the argument is a variable name and remove it.
// 		if err := s.backend.RemoveVariable(group, argument); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (s *Server) handleRequests(channel ssh.Channel, canWrite bool, raddr string, in <-chan *ssh.Request) {

// 	// Close the connection when we return.
// 	defer channel.Close()

// 	// Maintain a group state for this channel.
// 	environment := make(map[string]string)

// 	// Pull requests off the incoming channel.
// 	for request := range in {

// 		// Assume that this request is invalid.
// 		ok := false

// 		// Switch on the request type.
// 		switch request.Type {
// 		case "env":

// 			// Try to parse the payload. If we can't then there's nothing we
// 			// can do with this request.
// 			if payload, err := UnpackMessage(request.Payload); err != nil {

// 				// Write the error message to the log.
// 				log.Println(err)
// 			} else {

// 				// Write the payload slice into the environment map.
// 				for i := 0; i < len(payload)/2; i++ {
// 					environment[payload[i*2]] = payload[i*2+1]
// 				}

// 				// Indicate success.
// 				ok = true
// 			}

// 		case "exec":

// 			// Try to parse the payload. If we can't we don't want to
// 			// proceed.
// 			payload, err := UnpackMessage(request.Payload)
// 			if err != nil {

// 				// Notify the caller that we couldn't run the command.
// 				request.Reply(false, nil)
// 				return
// 			}

// 			// Indicate that we have started running the command.
// 			request.Reply(true, nil)

// 			// The exit status will be reported as a 4-byte, little-endian integer.
// 			exitStatusBuffer := bytes.NewBuffer([]byte{})

// 			// Run the command, reporting any error as a failure.
// 			if err := s.exec(channel, canWrite, raddr, environment, payload[0]); err != nil {

// 				// Write the error message to the log.
// 				log.Println(err)
// 				binary.Write(exitStatusBuffer, binary.BigEndian, uint32(1))
// 			} else {
// 				binary.Write(exitStatusBuffer, binary.BigEndian, uint32(0))
// 			}

// 			// Write the exit status.
// 			channel.SendRequest("exit-status", false, exitStatusBuffer.Bytes())

// 			// Only one exec command can be handled per channel, so we're done.
// 			return
// 		}

// 		// If requested, reply with the status.
// 		if request.WantReply {
// 			request.Reply(ok, nil)
// 		}
// 	}
// }

// func (s *Server) handleChannels(canWrite bool, raddr string, in <-chan ssh.NewChannel) {

// 	// Pull channels off the incoming channel.
// 	for newChannel := range in {

// 		// Only accept sessions.
// 		if newChannel.ChannelType() != "session" {
// 			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
// 			continue
// 		}

// 		// Attempt to accecpt the session channel.
// 		channel, requests, err := newChannel.Accept()
// 		if err != nil {
// 			log.Printf("server: could not accept channel: %s\n", err.Error())
// 			continue
// 		}

// 		go s.handleRequests(channel, canWrite, raddr, requests)
// 	}
// }

// // ListenAndServe starts a new SSH server listening on the given address.
// func (s *Server) ListenAndServe(address string) error {

// 	// Get the listeners lock.
// 	s.listenersMu.Lock()

// 	// Start listening.
// 	listener, err := net.Listen("tcp", address)
// 	if err != nil {
// 		s.listenersMu.Unlock()
// 		return err
// 	}

// 	// Add the listener to the list and release the lock.
// 	s.listeners.PushBack(listener)
// 	s.listenersMu.Unlock()

// 	for {

// 		// Accept a new connection.
// 		nConn, err := listener.Accept()
// 		if err != nil {
// 			return err
// 		}

// 		sConn, chans, reqs, err := ssh.NewServerConn(nConn, s.serverConfig)
// 		if err != nil {
// 			log.Println("server: failed to handshake")
// 			continue
// 		}

// 		// Determine whether or not the user can write.
// 		canWrite := false
// 		if sConn.User() == WriterUser {
// 			canWrite = true
// 		}
// 		// The incoming Request channel must be serviced.
// 		go ssh.DiscardRequests(reqs)

// 		// Service the incoming Channel channel.
// 		go s.handleChannels(canWrite, nConn.RemoteAddr().String(), chans)
// 	}

// 	return nil
// }

// func (s *Server) Stop() error {

// 	// Get the listeners lock and defer its closing.
// 	s.listenersMu.Lock()
// 	defer s.listenersMu.Unlock()

// 	for e := s.listeners.Front(); e != nil; e = e.Next() {

// 		if err := e.Value.(net.Listener).Close(); err != nil {
// 			return err
// 		}

// 		s.listeners.Remove(e)
// 	}

// 	return nil
// }

type ServerError struct {
	Err string
}

func (e ServerError) Error() string {
	return fmt.Sprintf("server: %s", e.Err)
}
