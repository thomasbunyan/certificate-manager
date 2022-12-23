package dns01

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/platform/wait"
	"github.com/pkg/errors"
	"github.com/thomasbunyan/certbot-lambda/internal/common"
)

// https://github.com/go-acme/lego/blob/master/providers/dns/route53/route53.go

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	HostedZoneID       string
	TTL                int
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	client *route53.Client
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for the AWS Route 53 service.
func Route53DNSProvider(client *route53.Client) challenge.Provider {
	return &DNSProvider{
		client: client,
		config: &Config{
			TTL:                10,
			PropagationTimeout: 2 * time.Minute,
			PollingInterval:    4 * time.Second,
		}}
}

// Present creates a TXT record using the specified parameters.
func (d *DNSProvider) Present(domain, token, keyAuth string) error {
	fqdn, value := dns01.GetRecord(domain, keyAuth)

	hostedZoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return errors.Wrap(err, "route53: failed to determine hosted zone ID")
	}

	records, err := d.getExistingRecordSets(hostedZoneID, fqdn)
	if err != nil {
		return errors.Wrap(err, "route53")
	}

	realValue := `"` + value + `"`

	var found bool
	for _, record := range records {
		if common.Val(record.Value) == realValue {
			found = true
		}
	}

	if !found {
		records = append(records, types.ResourceRecord{Value: common.Ptr(realValue)})
	}

	recordSet := &types.ResourceRecordSet{
		Name:            common.Ptr(fqdn),
		Type:            types.RRTypeTxt,
		TTL:             common.Ptr(int64(d.config.TTL)),
		ResourceRecords: records,
	}

	err = d.changeRecord(types.ChangeActionUpsert, hostedZoneID, recordSet)
	if err != nil {
		return errors.Wrap(err, "route53")
	}
	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(domain, token, keyAuth string) error {
	fqdn, _ := dns01.GetRecord(domain, keyAuth)

	hostedZoneID, err := d.getHostedZoneID(fqdn)
	if err != nil {
		return errors.Wrap(err, "route53: failed to determine Route 53 hosted zone ID")
	}

	records, err := d.getExistingRecordSets(hostedZoneID, fqdn)
	if err != nil {
		return errors.Wrap(err, "route53")
	}

	if len(records) == 0 {
		return nil
	}

	recordSet := &types.ResourceRecordSet{
		Name:            common.Ptr(fqdn),
		Type:            types.RRTypeTxt,
		TTL:             common.Ptr(int64(d.config.TTL)),
		ResourceRecords: records,
	}

	err = d.changeRecord(types.ChangeActionDelete, hostedZoneID, recordSet)
	if err != nil {
		return errors.Wrap(err, "route53")
	}
	return nil
}

func (d *DNSProvider) getHostedZoneID(fqdn string) (string, error) {
	if d.config.HostedZoneID != "" {
		return d.config.HostedZoneID, nil
	}

	authZone, err := dns01.FindZoneByFqdn(fqdn)
	if err != nil {
		return "", err
	}

	// .DNSName should not have a trailing dot
	reqParams := &route53.ListHostedZonesByNameInput{
		DNSName: common.Ptr(dns01.UnFqdn(authZone)),
	}
	res, err := d.client.ListHostedZonesByName(context.TODO(), reqParams)
	if err != nil {
		return "", err
	}

	var hostedZoneID string
	for _, hostedZone := range res.HostedZones {
		// .Name has a trailing dot
		if !hostedZone.Config.PrivateZone && common.Val(hostedZone.Name) == authZone {
			hostedZoneID = common.Val(hostedZone.Id)
			break
		}
	}

	if hostedZoneID == "" {
		return "", errors.Errorf("route53: zone %s not found for domain %s", authZone, fqdn)
	}

	hostedZoneID = strings.TrimPrefix(hostedZoneID, "/hostedzone/")

	return hostedZoneID, nil
}

func (d *DNSProvider) getExistingRecordSets(hostedZoneID string, fqdn string) ([]types.ResourceRecord, error) {
	listInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    common.Ptr(hostedZoneID),
		StartRecordName: common.Ptr(fqdn),
		StartRecordType: types.RRTypeTxt,
	}

	recordSetsOutput, err := d.client.ListResourceRecordSets(context.TODO(), listInput)
	if err != nil {
		return nil, err
	}

	if recordSetsOutput == nil {
		return nil, nil
	}

	var records []types.ResourceRecord
	for _, recordSet := range recordSetsOutput.ResourceRecordSets {
		if common.Val(recordSet.Name) == fqdn {
			records = append(records, recordSet.ResourceRecords...)
		}
	}

	return records, nil
}

func (d *DNSProvider) changeRecord(action types.ChangeAction, hostedZoneID string, recordSet *types.ResourceRecordSet) error {
	recordSetInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: common.Ptr(hostedZoneID),
		ChangeBatch: &types.ChangeBatch{
			Comment: common.Ptr("ACME DNS challenge"),
			Changes: []types.Change{{
				Action:            action,
				ResourceRecordSet: recordSet,
			}},
		},
	}

	res, err := d.client.ChangeResourceRecordSets(context.TODO(), recordSetInput)
	if err != nil {
		return errors.Wrap(err, "route53: failed to change record set")
	}

	changeID := res.ChangeInfo.Id

	return wait.For("route53", d.config.PropagationTimeout, d.config.PollingInterval, func() (bool, error) {
		reqParams := &route53.GetChangeInput{Id: changeID}

		res, err := d.client.GetChange(context.TODO(), reqParams)
		if err != nil {
			return false, errors.Wrap(err, "route53: failed to query change status")
		}

		if res.ChangeInfo.Status == types.ChangeStatusInsync {
			return true, nil
		}

		return false, errors.Errorf("route53: unable to retrieve change: ID=%s", common.Val(changeID))
	})
}
