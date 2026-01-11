package config

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type Postgres struct {
	DbName                      string
	DbUser                      string
	DatabaseInstanceType        string
	DatabaseBackupRetentionDays int
	DatabaseMultiAz             bool
}

type Config struct {
	ProjectName string
	Environment string
	VpcCidr     string

	HostedZoneName string
	Postgres       Postgres
}

func Must(ctx *pulumi.Context) Config {
	pulumiCfg := config.New(ctx, "")

	projectName := pulumiCfg.Require("projectName")
	environment := pulumiCfg.Require("environment")
	hostedZoneName := pulumiCfg.Require("hostedZoneName")
	vpcCidr := pulumiCfg.Require("vpcCidr")
	var postgres Postgres
	pulumiCfg.RequireObject("postgres", &postgres)

	cfg := Config{
		ProjectName:    projectName,
		Environment:    environment,
		HostedZoneName: hostedZoneName,
		Postgres:       postgres,
		VpcCidr:        vpcCidr,
	}
	return cfg
}
