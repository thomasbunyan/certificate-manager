package acme

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"github.com/pkg/errors"
)

// Obtain returns a certificate using a DNS-01 challenge provider for a given domain list
func Obtain(p challenge.Provider, domains []string, email string) (*certificate.Resource, error) {
	privateKey, err := generatePrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "obtain: failed to generate private key")
	}

	account := &Account{
		Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(account)
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "obtain: failed to initialize client")
	}

	// Set up challenge
	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		return nil, errors.Wrap(err, "obtain: failed to set provider")
	}

	// Register
	registration, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, errors.Wrap(err, "obtain: failed to register account")
	}
	account.Registration = registration

	// Request certificate
	certificate, err := client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "obtain: failed to obtain certificate")
	}

	return certificate, nil
}

// generatePrivateKey generates and returns an RSA private key
func generatePrivateKey() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return key, nil
}
