package ssh

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

const (
	User = `context`
)

type Server struct {
	Addr      string
	HostKey   ssh.Signer
	Commands  map[string]CommandFactory
	Authorize func(net.Addr, ssh.PublicKey) ([]string, error)
}

// ListenAndServe starts a new SSH server listening on the given address.
func (s *Server) ListenAndServe() error {

	serverConfig := &ssh.ServerConfig{
		Config: ssh.Config{
			Ciphers: []string{"aes256-ctr"},
			MACs:    []string{"hmac-sha1"},
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {

			// Insure that the correct user has been specified.
			if conn.User() != User {
				return nil, ServerError{"unauthorized"}
			}

			commands, err := s.Commands(conn.RemoteAddr(), key)
			if err != nil {
				return nil, err
			}

			permissions := &ssh.Permissions{Extensions: make(map[string]string)}
			for _, command := range commands {
				permissions.Extensions[command] = ""
			}

			return permissions, nil
		},
	}
	serverConfig.AddHostKey(s.HostKey)

	// Start listening.
	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		fmt.Println("listner err", s.Addr)
		return err
	}

	// Connection loop.
	for {

		// Accept a new connection.
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return err
		}

		// Handle the connection.
		go s.handleConn(conn)
	}

	return nil
}

func (s *Server) handleConn(conn net.Conn) {

	serverConn, newChannels, reqs, err := ssh.NewServerConn(conn, serverConfig)
	if err != nil {
		log.Println("server: failed to handshake")
		return
	}

	// The incoming Request channel must be serviced, but shouldn't contain
	// any requests we care about.
	go ssh.DiscardRequests(reqs)

	// Iterate through the incomming channel requests.
	var wait sync.WaitGroup
	for newChannel := range newChannels {
		wait.Add(1)
		go func() {
			handleNewChannel(channel, serverConn.Permissions)
			wait.Done()
		}()
	}

	wait.Wait()
	if err := serverConn.Close(); err != nil {
		log.Println(err)
	}
}

func (s *Server) handleNewChannel(newChannel ssh.NewChannel, permissions *ssh.Permissions) {

	// Only accept sessions.
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return
	}

	// Attempt to accecpt the session channel.
	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("server: could not accept channel: %s\n", err.Error())
		return
	}

	defer func() {
		if err := channel.Close(); err != nil {
			log.Println(err)
		}
	}()

	// Create an environment specific to the channel.
	env := make(map[string]string)

	// Iterate through the requests on the channel.
	for request := range requests {

		// Unpack the payload.
		payload, err := UnpackMessage(request.Payload)
		if err != nil {
			log.Println(err)
			return
		}

		// Switch on the request type.
		switch request.Type {
		case "env":
			for i := 0; i < len(payload)/2; i++ {
				env[payload[i*2]] = payload[i*2+1]
			}

		case "exec":

			// Parse the command a if it was on the command line.
			commandArgs := strings.SplitN(payload[0], ` `, 2)
			command := commandArgs[0]

			// Check if this command is allowed.
			if _, ok := permissions.Extensions[command]; !ok {
				request.Reply(false, nil)
				return
			}

			commandFactory := s.Commands[command]
			if commandFactory == nil {
				request.Reply(false, nil)
				return
			}

			// Make sure the command is allowed, and that we have a factory
			// for it.
			if commandAllowed == nil || commandFactory == nil {
				request.Reply(false, nil)
			} else {

				// Indicate that we have started running the command.
				request.Reply(true, nil)

				// The exit status will be reported as a 4-byte, little-endian integer.
				exitStatusBuffer := bytes.NewBuffer([]byte{})

				// Get a new command object of the

				command, err := commandFactory()

				comm := commandFactory.Run(commandArgs, env)

				// Run the command, reporting any error as a failure.
				if err := s.Handler.ServeSSH(channel, env, command); err != nil {
					log.Println(err)
					binary.Write(exitStatusBuffer, binary.BigEndian, uint32(1))
				} else {
					binary.Write(exitStatusBuffer, binary.BigEndian, uint32(0))
				}

				// Write the exit status and close the channel. Only
				// one exec command can be run per channel.
				channel.SendRequest("exit-status", false, exitStatusBuffer.Bytes())
			}

			// Only one exec command can be handled per channel, so we're done.

		}
	}

}

type ServerError struct {
	Err string
}

func (e ServerError) Error() string {
	return fmt.Sprintf("server: %s", e.Err)
}
