package certlambda

import "testing"

func TestCertLambda(t *testing.T) {
	input := Event{
		DomainNames:    []string{},
		Email:          "foo",
		RenewThreshold: 7,
	}
	Run(input)
}
