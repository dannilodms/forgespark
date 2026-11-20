package main

import (
	"crypto/ed25519"
	"encoding/pem"
	"strings"

	"golang.org/x/crypto/ssh"
)

// serverKeys are the two key pairs a new server is made with, both drawn
// anew on each turn-on: the deploy key logs in as the deploy user, and the
// host key is the one the server answers SSH with, known before the first
// connection.
type serverKeys struct {
	// deployPrivate is the private deploy key in OpenSSH's format, the one
	// the local copy keeps.
	deployPrivate []byte
	// deployPublic is the public deploy key as a line of authorized keys.
	deployPublic string
	deploySigner ssh.Signer
	// hostPrivate is the private host key in OpenSSH's format, the one the
	// cloud-init file gives the server.
	hostPrivate string
	hostPublic  ssh.PublicKey
}

func newServerKeys(name string) (serverKeys, error) {
	_, deploy, err := ed25519.GenerateKey(nil)
	if err != nil {
		return serverKeys{}, err
	}
	_, host, err := ed25519.GenerateKey(nil)
	if err != nil {
		return serverKeys{}, err
	}
	deployPrivate, err := ssh.MarshalPrivateKey(deploy, "deploy@"+name)
	if err != nil {
		return serverKeys{}, err
	}
	hostPrivate, err := ssh.MarshalPrivateKey(host, "")
	if err != nil {
		return serverKeys{}, err
	}
	deploySigner, err := ssh.NewSignerFromKey(deploy)
	if err != nil {
		return serverKeys{}, err
	}
	hostPublic, err := ssh.NewPublicKey(host.Public())
	if err != nil {
		return serverKeys{}, err
	}
	return serverKeys{
		deployPrivate: pem.EncodeToMemory(deployPrivate),
		deployPublic:  authorizedLine(deploySigner.PublicKey()),
		deploySigner:  deploySigner,
		hostPrivate:   string(pem.EncodeToMemory(hostPrivate)),
		hostPublic:    hostPublic,
	}, nil
}

// login is what logs in to the server made with these keys.
func (k serverKeys) login() loginKeys {
	return loginKeys{deploy: k.deploySigner, host: k.hostPublic}
}

// loginKeys log in to a server: the deploy key, and the only host key
// trusted.
type loginKeys struct {
	deploy ssh.Signer
	host   ssh.PublicKey
}

// authorizedLine writes a public key as a line of authorized keys, with no
// line break.
func authorizedLine(key ssh.PublicKey) string {
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
}
