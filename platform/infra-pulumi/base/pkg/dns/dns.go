package dns

import (
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/acm"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strings"
)

type DnsRecords struct {
	Environment     string
	HostedZoneName  string // "fs0ciety.dev"
	IncludeWildcard bool   // optional: if you want *.fs0ciety.dev too
}

type Result struct {
	HostedZoneId pulumi.StringOutput
	CertArn      pulumi.StringOutput
}

func NewDnsRecords(ctx *pulumi.Context, projectName string, d DnsRecords) (*Result, error) {
	zone, err := route53.LookupZone(ctx, &route53.LookupZoneArgs{
		Name: pulumi.StringRef(d.HostedZoneName),
	}, nil)
	if err != nil {
		return nil, err
	}

	// --- Domains we want the cert to cover ---
	apex := d.HostedZoneName         // "fs0ciety.dev"
	www := "www." + d.HostedZoneName // "www.fs0ciety.dev"
	wild := "*." + d.HostedZoneName  // "*.fs0ciety.dev" (optional)

	domains := []string{apex, www}
	if d.IncludeWildcard {
		domains = append(domains, wild)
	}

	name := fmt.Sprintf("%s-%s-cert", projectName, d.Environment)

	// Primary is apex; SANs are the rest
	sans := pulumi.StringArray{}
	for _, s := range domains[1:] {
		sans = append(sans, pulumi.String(s))
	}

	cert, err := acm.NewCertificate(ctx, name, &acm.CertificateArgs{
		DomainName:              pulumi.String(domains[0]), // ✅ apex
		SubjectAlternativeNames: sans,                      // ✅ www (+ optional wildcard)
		ValidationMethod:        pulumi.String("DNS"),
		Tags: pulumi.StringMap{
			"Environment": pulumi.String(d.Environment),
			"Name":        pulumi.String(name),
		},
	})
	if err != nil {
		return nil, err
	}

	// --- Create one validation record per domain (WITHOUT relying on ordering) ---
	validationFqdns := pulumi.StringArray{}

	// helper: find validation option for a given domain
	find := func(domain string, opts []acm.CertificateDomainValidationOption) (name, typ, value string) {
		for _, o := range opts {
			if o.DomainName != nil && *o.DomainName == domain {
				if o.ResourceRecordName != nil {
					name = *o.ResourceRecordName
				}
				if o.ResourceRecordType != nil {
					typ = *o.ResourceRecordType
				}
				if o.ResourceRecordValue != nil {
					value = *o.ResourceRecordValue
				}
				return
			}
		}
		return
	}

	for _, domain := range domains {
		// compute rr fields by searching all validation options for this domain
		rrName := cert.DomainValidationOptions.ApplyT(func(opts []acm.CertificateDomainValidationOption) (string, error) {
			n, _, _ := find(domain, opts)
			if n == "" {
				return "", fmt.Errorf("acm validation record name missing for %s", domain)
			}
			return n, nil
		}).(pulumi.StringOutput)

		rrType := cert.DomainValidationOptions.ApplyT(func(opts []acm.CertificateDomainValidationOption) string {
			_, t, _ := find(domain, opts)
			return t
		}).(pulumi.StringOutput)

		rrValue := cert.DomainValidationOptions.ApplyT(func(opts []acm.CertificateDomainValidationOption) string {
			_, _, v := find(domain, opts)
			return v
		}).(pulumi.StringOutput)

		safeDomain := strings.NewReplacer(".", "-", "*", "wildcard").Replace(domain)
		recName := fmt.Sprintf("%s-%s-cert-val-%s", projectName, d.Environment, safeDomain)
		rec, err := route53.NewRecord(ctx, recName, &route53.RecordArgs{
			ZoneId:         pulumi.String(zone.ZoneId),
			Name:           rrName,
			Type:           rrType,
			AllowOverwrite: pulumi.Bool(true),
			Ttl:            pulumi.Int(60),
			Records: pulumi.StringArray{
				rrValue,
			},
		})
		if err != nil {
			return nil, err
		}

		validationFqdns = append(validationFqdns, rec.Fqdn)
	}

	// --- Wait until ACM is actually issued ---
	validationName := fmt.Sprintf("%s-%s-cert-validation", projectName, d.Environment)
	validation, err := acm.NewCertificateValidation(ctx, validationName, &acm.CertificateValidationArgs{
		CertificateArn:        cert.Arn,
		ValidationRecordFqdns: validationFqdns,
	})
	if err != nil {
		return nil, err
	}

	return &Result{
		HostedZoneId: pulumi.String(zone.ZoneId).ToStringOutput(),
		CertArn:      validation.CertificateArn,
	}, nil
}
