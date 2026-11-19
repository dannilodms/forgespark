// Package smtptest gives a test an SMTP server that refuses every recipient
// and repeats the address in its answer, as many servers do. A test sends to
// it to prove that a failed sending carries the address nowhere: not in an
// error, a log, a span or the errors River keeps for a job.
package smtptest

import (
	"bufio"
	"net"
	"strings"
	"testing"
)

// RefusingServer is an SMTP server that refuses every recipient.
type RefusingServer struct {
	// Host and Port are where it listens, on the loopback interface.
	Host string
	Port int
	// Refused receives a value each time a recipient is refused.
	Refused <-chan struct{}

	refused  chan struct{}
	listener net.Listener
}

// NewRefusingServer starts a server that lives until the test ends.
func NewRefusingServer(t testing.TB) *RefusingServer {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for SMTP: %v", err)
	}
	address, _ := listener.Addr().(*net.TCPAddr)
	refused := make(chan struct{}, 64)
	server := &RefusingServer{Host: "127.0.0.1", Port: address.Port, Refused: refused, refused: refused, listener: listener}
	t.Cleanup(func() { listener.Close() })
	go server.accept()
	return server
}

// Address is where the server listens, as host:port.
func (s *RefusingServer) Address() string {
	return s.listener.Addr().String()
}

// accept takes clients until the listener is closed.
func (s *RefusingServer) accept() {
	for connection, err := s.listener.Accept(); err == nil; connection, err = s.listener.Accept() {
		go s.converse(connection)
	}
}

// converse answers one client until it quits or goes away.
func (s *RefusingServer) converse(connection net.Conn) {
	defer connection.Close()
	reader := bufio.NewReader(connection)
	answer(connection, "220 refusing.test ESMTP")
	for open := true; open; {
		line, err := reader.ReadString('\n')
		open = err == nil && s.answer(connection, strings.TrimRight(line, "\r\n"))
	}
}

// answer answers one command, and tells whether the conversation goes on.
func (s *RefusingServer) answer(connection net.Conn, command string) bool {
	verb := strings.ToUpper(strings.SplitN(command, " ", 2)[0])
	switch verb {
	case "EHLO":
		answer(connection, "250-refusing.test\r\n250-ENHANCEDSTATUSCODES\r\n250 8BITMIME")
	case "HELO", "NOOP", "RSET":
		answer(connection, "250 2.0.0 Ok")
	case "MAIL":
		answer(connection, "250 2.1.0 Ok")
	case "RCPT":
		s.refuse(connection, command)
	case "QUIT":
		answer(connection, "221 2.0.0 Bye")
		return false
	default:
		answer(connection, "502 5.5.2 Error: command not recognized")
	}
	return true
}

// refuse refuses the recipient of `RCPT TO:<address>`, repeating the
// address, as a server that knows no such mailbox does.
func (s *RefusingServer) refuse(connection net.Conn, command string) {
	_, address, _ := strings.Cut(command, ":")
	answer(connection, "550 5.1.1 "+strings.TrimSpace(address)+": Recipient address rejected: User unknown")
	select {
	case s.refused <- struct{}{}:
	default:
		// Nobody reads them: the test already knows a recipient was refused.
	}
}

func answer(connection net.Conn, line string) {
	_, _ = connection.Write([]byte(line + "\r\n"))
}
