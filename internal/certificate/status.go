package certificate

import (
	"math"
	"time"
)

type RenewalStatus struct {
	// A Time object that represents the expiration date for the certificate.
	NotAfter time.Time

	// The number of days until the certificate expires.
	DaysRemaining int16

	// The number of days until the certificate is due to renewal.
	DaysUntilRenewal int16

	// Flag denoting whether the certificate is valid for renewal.
	RenewalValid bool
}

func CheckExpiration(threshold uint16, notAfter time.Time) RenewalStatus {
	daysRemaining := int16(math.Floor(notAfter.Sub(time.Now().In(time.UTC)).Hours() / 24))
	daysUntilRenewal := daysRemaining - int16(threshold)
	return RenewalStatus{
		NotAfter:         notAfter,
		DaysRemaining:    daysRemaining,
		DaysUntilRenewal: daysUntilRenewal,
		RenewalValid:     daysUntilRenewal < 1,
	}
}
