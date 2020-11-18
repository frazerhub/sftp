// Package sftp is a small client for managing file uploads to an sftp server.
// The surface area has been kept intentionally small to limit potentially
// dangerous operations, but should include all of the functionality we require.
//
// The primary use of this package is for doing vehicle uploads to our
// third-party vendors that do not have an API, though if you are working with
// a partner that still uses raw FTP, you will need to use a different package.
package sftp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Config is the required configuration for initializing a new Client.
type Config struct {
	User     string // The username for connecting to the server
	Password string // The passwword for connecting to the server
	Addr     string // The address of the server, including port e.g. my.ftp.com:22
}

// NewClient initializes a new Client by connecting via SSH and establishing a
// client connection. This process can fail if either the SSH or SFTP
// connections fail. If it succeeds, the returned Client is safe to call from
// multiple goroutines.
func NewClient(cfg Config) (*Client, error) {
	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshClient, err := ssh.Dial("tcp", cfg.Addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect ssh: %w", err)
	}
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("failed to connect sftp: %w", err)
	}
	return &Client{
		sshClient:  sshClient,
		sftpClient: sftpClient,
	}, nil
}

// Client provides methods for managing files on a remote SFTP server.
type Client struct {
	// We hold a handle to the underlying ssh client in case we want to add
	// reconnect behavior in the future.
	sshClient *ssh.Client

	// We wrap the sftp.Client because there are some oddities especially around
	// creating files, where the calling Open fails against the AWS Transfer
	// Family, as it attempts to create the file with read permissions. This may
	// not be an issue with other SFTP servers, but this client should work
	// regardless, as it simply uses a smaller subset of permissions.
	sftpClient *sftp.Client
}

// Close closes the underlying sftp connection.
func (c Client) Close() error {
	return c.sftpClient.Close()
}

// Open attempts to read the file at path and returns an *sftp.File that
// implements io.Reader, io.Writer, and io.Closer.
func (c Client) Open(_ context.Context, path string) (*sftp.File, error) {
	f, err := c.sftpClient.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}
	return f, nil
}

// ReadDir attempts to read the directory at path and returns a slice of
// os.FileInfos that can be used to get additional information about the files.
func (c Client) ReadDir(_ context.Context, path string) ([]os.FileInfo, error) {
	files, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}
	return files, nil
}

// Remove attempts to delete the file at path.
func (c Client) Remove(_ context.Context, path string) error {
	if err := c.sftpClient.Remove(path); err != nil {
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}
	return nil
}

// Save creates a new file with the given filename and writes the contents of r
// to it.
func (c Client) Save(_ context.Context, filename string, r io.Reader) error {
	f, err := c.sftpClient.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to save file %s: %w", filename, err)
	}

	return nil
}

// SaveBytes is a convenience wrapper around Save that wraps the provided bytes
// in an io.Reader.
func (c Client) SaveBytes(ctx context.Context, filename string, bs []byte) error {
	return c.Save(ctx, filename, bytes.NewReader(bs))
}

// SaveString is a convenience wrapper around Save that wraps the provided
// string in an io.Reader.
func (c Client) SaveString(ctx context.Context, filename, s string) error {
	return c.Save(ctx, filename, strings.NewReader(s))
}
