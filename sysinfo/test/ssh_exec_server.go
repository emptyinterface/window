package test

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

type (
	SSHExecServer struct {
		Host string
		l    net.Listener
	}
)

var (
	signer ssh.Signer
)

func init() {
	const privateBytes = `-----BEGIN RSA PRIVATE KEY-----
MIIBzAIBAAJhAOIAwMVZCOtUEjtrGsv0CkDTYgGGeS4z5sgtaTrwg/6gWYMtTSWc
zgQ9wmpdo2rNZypUUXy2cXzAyiaUwp4jXSctPYVYErLk0KGycK6SaJogu7HAemiZ
3TLn8QkfODbakQIDAQABAmEA4PDY7VNx0jAKOYOf1zGdZuo9mMEMKdVUtRalrxkm
dy+ICEz1hSMt1gDWWWG7vhiS4ALlW/TKFMP6E4rkiqG+tQ3thrdEwyeFFQBzBoyq
dhb7Dgipez5ELh3282g8dWsxAjEA98oYjJ4Gds7gCFenc8daNxdhSdKu3GVY32kV
aV8/Quhpq2lTywYlsvRs7bN6u3WtAjEA6X3ZuxGt55h2AHhwO9mzU9DS3KPP15iA
i0zieVb/Tg3i/iykHy5kkRzzuujQm6z1AjEA5NT7ROkvGQtF9A5W82I4G0Z5Lz7l
A16I65FVF8HBX13ZMFaN7qGXsSNvcTld777lAjEA5O1/jOrIl0nkaJGteQD50jPs
imgSYFAluG6pnk6uAtmatZsPT4MtFxpL3fZmkjwBAjAZA2joUKAAW//N3zHzlZNO
6CaM8izhmFh3Bn1KM1ByPzpoHcvIzScvS9f4j9iVOMw=
-----END RSA PRIVATE KEY-----`
	var err error
	signer, err = ssh.ParsePrivateKey([]byte(privateBytes))
	if err != nil {
		log.Fatal(err)
	}
}

// func NewPayloadSSHHandler(t *testing.T, response []byte) func(req *ssh.Request, channel ssh.Channel) {

// 	return func(req *ssh.Request, channel ssh.Channel) {

// 		gzw := gzip.NewWriter(channel)

// 		if n, err := gzw.Write([]byte(MetricsResponsePayload)); err != nil {
// 			t.Error(err)
// 		} else if n < len(MetricsResponsePayload) {
// 			t.Errorf("Short Write! expected %d bytes, got %d", n, len(MetricsResponsePayload))
// 		}
// 		gzw.Close()
// 		if _, err := channel.SendRequest("exit-status", false, []byte{0, 0, 0, 0}); err != nil {
// 			t.Error(err)
// 		}
// 		if err := channel.Close(); err != nil {
// 			t.Error(err)
// 		}

// 	}

// }

func NewTestSSHExecServer(t *testing.T, f func(req *ssh.Request, channel ssh.Channel)) *SSHExecServer {

	l, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Error(err)
		t.Fatal("failed to acquire tcp listener")
	}

	go func() {

		for {
			conn, err := l.Accept()
			if err != nil {
				// closing the listener throws an error, we can safely ignore it
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					return
				}
				t.Error(err)
				t.Fatal("failed to accept incoming connection")
			}
			go func(conn net.Conn) {
				defer conn.Close()

				// Before use, a handshake must be performed on the incoming
				config := &ssh.ServerConfig{NoClientAuth: true}
				config.AddHostKey(signer)
				_, chans, reqs, err := ssh.NewServerConn(conn, config)
				if err != nil {
					t.Error(err)
					t.Fatal("failed to handshake")
				}

				// The incoming Request channel must be serviced.
				go ssh.DiscardRequests(reqs)

				for c := range chans {

					if t := c.ChannelType(); t != "session" {
						c.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
						continue
					}

					channel, requests, err := c.Accept()
					if err != nil {
						t.Error(err)
						t.Fatal("could not accept channel.")
					}

					switch req := <-requests; req.Type {
					case "exec":
						req.Reply(true, nil)
						req.Payload, err = parse_command(req.Payload)
						if err != nil {
							t.Error(err)
						}
						f(req, channel)
					default:
						req.Reply(false, nil)
					}

					channel.Close()

				}

			}(conn)
		}

	}()

	return &SSHExecServer{
		Host: l.Addr().String(),
		l:    l,
	}

}

func (s *SSHExecServer) Close() error {
	return s.l.Close()
}

func parse_command(payload []byte) ([]byte, error) {

	if len(payload) == 0 {
		return nil, fmt.Errorf("unable to parse command: empty payload supplied")
	}

	length := int(binary.BigEndian.Uint32(payload))
	payload = payload[4:]

	if length > len(payload) {
		return nil, fmt.Errorf("specified length (%d) longer than payload length (%d)", length, payload)
	}

	return payload[:length], nil

}
