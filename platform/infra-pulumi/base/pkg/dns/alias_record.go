package dns

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type AliasARecord struct {
	Environment      string
	Service          string
	Route53ZoneId    pulumi.StringInput
	RecordDomainName string
	AlbZoneId        pulumi.StringInput
	AlbDnsName       pulumi.StringInput
}

func NewAliasARecord(ctx *pulumi.Context, projectName string, a AliasARecord) (*route53.Record, error) {
	name := fmt.Sprintf("%s-%s-%s-alias-record", projectName, a.Environment, a.Service)
	record, err := route53.NewRecord(ctx, name, &route53.RecordArgs{
		ZoneId: a.Route53ZoneId,
		Name:   pulumi.String(a.RecordDomainName),
		Type:   pulumi.String("A"),
		Aliases: route53.RecordAliasArray{
			&route53.RecordAliasArgs{
				Name:                 a.AlbDnsName,
				ZoneId:               a.AlbZoneId,
				EvaluateTargetHealth: pulumi.Bool(true),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}
