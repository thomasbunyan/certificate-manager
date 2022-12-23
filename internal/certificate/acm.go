package certificate

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/pkg/errors"
)

// Returns the ACM certificate summary for a given domain, if it exists.
func GetCertificateSummary(client *acm.Client, domains *[]string) (*types.CertificateSummary, error) {
	res, err := client.ListCertificates(context.TODO(), &acm.ListCertificatesInput{})
	if err != nil {
		return nil, err
	}

	for _, cert := range res.CertificateSummaryList {
		if contains(domains, cert.DomainName) {
			return &cert, nil
		}
	}

	return nil, nil
}

// Returns the CertificateDetail for a given ACM certificate ARN.
func GetCertificateDetails(client *acm.Client, arn *string) (*types.CertificateDetail, error) {
	res, err := client.DescribeCertificate(context.TODO(), &acm.DescribeCertificateInput{CertificateArn: arn})
	if err != nil {
		return nil, err
	}
	return res.Certificate, nil
}

// Updates a given certificate in ACM with a new provided cert.
func ImportCertificate(client *acm.Client, arn *string, cert *certificate.Resource) error {
	serverCertificate, err := retrieveServerCertificate(cert.Certificate)
	if err != nil {
		return errors.Wrap(err, "acm: unable to retrieve server certificate")
	}

	_, err = client.ImportCertificate(context.TODO(), &acm.ImportCertificateInput{
		CertificateArn:   arn,
		Certificate:      serverCertificate,
		PrivateKey:       cert.PrivateKey,
		CertificateChain: cert.IssuerCertificate,
	})
	return err
}

// contains checks whether a string slice contains a given string
func contains(s *[]string, str *string) bool {
	for _, v := range *s {
		if v == *str {
			return true
		}
	}

	return false
}

// retrieveServerCertificate retrieves the server certificate from the given PEM encoded list
func retrieveServerCertificate(list []byte) ([]byte, error) {
	var blocks []*pem.Block
	for {
		var certDERBlock *pem.Block
		certDERBlock, list = pem.Decode(list)
		if certDERBlock == nil {
			break
		}

		if certDERBlock.Type == "CERTIFICATE" {
			blocks = append(blocks, certDERBlock)
		}
	}

	crt := bytes.NewBuffer(nil)
	for _, block := range blocks {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse certificate")
		}

		if !cert.IsCA {
			pem.Encode(crt, block)
			break
		}
	}

	return crt.Bytes(), nil
}
