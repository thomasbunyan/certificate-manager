package certlambda

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/pkg/errors"
	"github.com/thomasbunyan/certbot-lambda/internal/acme"
	"github.com/thomasbunyan/certbot-lambda/internal/acme/challenge/dns01"
	"github.com/thomasbunyan/certbot-lambda/internal/certificate"

	cert "github.com/go-acme/lego/v4/certificate"
)

var (
	acmClient     *acm.Client
	route53Client *route53.Client
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Panicf("Failed loading configuration for AWS SDK: %v", err)
	}

	acmClient = acm.NewFromConfig(cfg, func(o *acm.Options) {
		o.Region = "us-east-1"
	})

	route53Client = route53.NewFromConfig(cfg, func(o *route53.Options) {})
}

func Run(event Event) error {
	log.Printf("certlambda: Starting execution: %+v", event)

	certSummary, err := certificate.GetCertificateSummary(acmClient, &event.DomainNames)
	if err != nil {
		return errors.Wrap(err, "certlambda: Failed listing ACM certificates")
	}

	if certSummary == nil {
		log.Printf("certlambda: No active certificate found for domain(s) %v", event.DomainNames)
		log.Println("certlambda: Requesting new certificate")
		newCert, err := acme.Obtain(dns01.Route53DNSProvider(route53Client), event.DomainNames, event.Email)
		if err != nil {
			return errors.Wrap(err, "certlambda: Failed to request new certificate")
		}

		log.Println("certlambda: Successfully requested new certificate")

		err = importCertificate(nil, newCert)
		if err != nil {
			return errors.Wrap(err, "certlambda: Unable to import certificate")
		}
	} else {
		log.Printf("certlambda: Certificate found for domain %v: %v", event.DomainNames, *certSummary.CertificateArn)

		certDetails, err := certificate.GetCertificateDetails(acmClient, certSummary.CertificateArn)
		if err != nil {
			return errors.Wrap(err, "certlambda: Failed to get certificate details")
		}

		renewalStatus := certificate.CheckExpiration(event.RenewThreshold, *certDetails.NotAfter)
		log.Printf("certlambda: Renewal status: %+v", renewalStatus)

		if renewalStatus.RenewalValid {
			log.Println("certlambda: Renewing certificate")
			newCert, err := acme.Obtain(dns01.Route53DNSProvider(route53Client), event.DomainNames, event.Email)
			if err != nil {
				return errors.Wrap(err, "certlambda: Failed to renew certificate")
			}

			log.Println("certlambda: Successfully renewed certificate")

			err = importCertificate(certDetails.CertificateArn, newCert)
			if err != nil {
				return errors.Wrap(err, "certlambda: Unable to import certificate")
			}
		} else {
			log.Println("certlambda: Certificate still valid, no action required")
		}
	}

	return nil
}

func importCertificate(certArn *string, cert *cert.Resource) error {
	log.Println("certlambda[import]: Importing certificate into ACM")

	err := certificate.ImportCertificate(acmClient, certArn, cert)
	if err != nil {
		return errors.Wrap(err, "certlambda[import]: Failed to import certificate into ACM")
	}

	log.Println("certlambda[import]: Successfully imported certificate")

	return nil
}
