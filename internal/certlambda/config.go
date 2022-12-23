package certlambda

type Event struct {
	// Fully qualified domain name, such as www.example.com or example.com, for the certificate.
	DomainNames []string `json:"domainName"`

	// Email used for registration and recovery contact.
	Email string `json:"email"`

	// Number of days before the certificate expiry until it should be renewed.
	RenewThreshold uint16 `json:"renewThreshold"`
}
