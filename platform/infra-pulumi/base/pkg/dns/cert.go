package dns

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Cert struct {
	Environment    string
	HostedZoneName string
	AppDomainName  string
}

// Creates an ACM certificate validated via Route53 DNS in the hosted zone.
func NewCert(ctx *pulumi.Context, name string, cert Cert) (*route53.Record, error) {
	zone, err := route53.LookupZone(ctx, &route53.LookupZoneArgs{
		Name: pulumi.StringRef(cert.HostedZoneName),
	}, nil)
	if err != nil {
		return nil, err
	}

	certRes, err := acm.NewCertificate(ctx, name+"-cert", &acm.CertificateArgs{
		DomainName:       pulumi.String(cert.AppDomainName),
		ValidationMethod: pulumi.String("DNS"),
		Tags: pulumi.StringMap{
			"Environment": pulumi.String(cert.Environment),
			"Name":        pulumi.String(fmt.Sprintf("traba-%s-cert", cert.Environment)),
		},
	})
	if err != nil {
		return nil, err
	}

	// Use the first validation option (common case: single domain).
	valName := certRes.DomainValidationOptions.Index(pulumi.Int(0)).ApplyT(func(v acm.CertificateDomainValidationOption) (string, error) {
		return *v.ResourceRecordName, nil
	}).(pulumi.StringOutput)
	valType := certRes.DomainValidationOptions.Index(pulumi.Int(0)).ResourceRecordType()
	valValue := certRes.DomainValidationOptions.Index(pulumi.Int(0)).ResourceRecordValue()
	recRes, err := route53.NewRecord(ctx, name+"-cert-validation", &route53.RecordArgs{
		ZoneId: pulumi.String(zone.ZoneId),
		Name:   valName,
		Type:   valType.Elem().ToStringOutput(),
		Ttl:    pulumi.Int(60),
		Records: pulumi.StringArray{
			valValue.Elem().ToStringOutput(),
		},
	})

	if err != nil {
		return nil, err
	}
	return recRes, nil
}
