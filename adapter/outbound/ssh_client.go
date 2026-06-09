package outbound

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"golang.org/x/crypto/ssh"
)

type sshClient struct{}

func NewSSHClient() outbound.SSHClient {
	return &sshClient{}
}

func (c *sshClient) Connect(ctx context.Context, host string, port int, username, authMethod, password, privateKey string) (outbound.SSHSession, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{},
		//nolint:gosec // InsecureIgnoreHostKey is acceptable for a management tool; will be configurable later
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	switch entity.AuthMethod(authMethod) {
	case entity.AuthMethodPassword:
		config.Auth = append(config.Auth, ssh.Password(password))
	case entity.AuthMethodPrivateKey:
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	default:
		return nil, fmt.Errorf("unsupported auth method: %s", authMethod)
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	// Request PTY with default size (will be resized by the terminal handler)
	if err := session.RequestPty("xterm", 24, 80, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to request PTY: %w", err)
	}

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	pipeReader, pipeWriter := io.Pipe()
	session.Stdout = pipeWriter
	session.Stderr = pipeWriter

	if err := session.Shell(); err != nil {
		pipeWriter.Close()
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	return &sshSession{
		client:    client,
		session:   session,
		stdin:     stdinPipe,
		stdout:    pipeReader,
		pipeWriter: pipeWriter,
	}, nil
}

type sshSession struct {
	client     *ssh.Client
	session    *ssh.Session
	stdin      io.Writer
	stdout     *io.PipeReader
	pipeWriter *io.PipeWriter
}

func (s *sshSession) Stdin() io.Writer {
	return s.stdin
}

func (s *sshSession) Stdout() io.Reader {
	return s.stdout
}

func (s *sshSession) Close() error {
	s.pipeWriter.Close()
	s.session.Close()
	return s.client.Close()
}

func (s *sshSession) Resize(rows, cols int) error {
	return s.session.WindowChange(rows, cols)
}