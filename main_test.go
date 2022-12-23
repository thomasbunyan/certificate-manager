package main

import (
	"testing"

	"github.com/thomasbunyan/certificate-manager/internal/certlambda"
)

func TestHandleRequest(t *testing.T) {
	input := certlambda.Event{
		DomainNames:    []string{""},
		Email:          "",
		RenewThreshold: 1,
	}
	HandleRequest(input)
}
