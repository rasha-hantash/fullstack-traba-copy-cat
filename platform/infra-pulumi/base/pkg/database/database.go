package database

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/rds"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Postgres struct {
	Environment      string
	PrivateSubnetIds pulumi.StringArrayInput
	DbSgId           pulumi.StringInput

	DbName         string
	MasterUsername string

	InstanceClass   string
	MultiAz         bool
	BackupRetention int
}

type Result struct {
	Instance *rds.Instance

	Endpoint pulumi.StringOutput
	Port     pulumi.IntOutput

	// When ManageMasterUserPassword=true, RDS creates a Secrets Manager secret
	// and exposes its ARN here.
	MasterUserSecretArn pulumi.StringPtrOutput
}

func NewPostgres(ctx *pulumi.Context, projectName string, p Postgres, opts ...pulumi.ResourceOption) (*Result, error) {
	name := fmt.Sprintf("%s-%s-db-subnet-group", projectName, p.Environment)
	subnetGroup, err := rds.NewSubnetGroup(ctx, name, &rds.SubnetGroupArgs{
		SubnetIds: p.PrivateSubnetIds,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(p.Environment),
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	// ✅ Key change: ManageMasterUserPassword=true
	// This tells AWS to generate and store the master password in Secrets Manager.
	name = fmt.Sprintf("%s-%s-postgres", projectName, p.Environment)
	db, err := rds.NewInstance(ctx, name, &rds.InstanceArgs{
		Identifier: pulumi.String(name),

		Engine:        pulumi.String("postgres"),
		EngineVersion: pulumi.String("18.1"),

		InstanceClass:    pulumi.String(p.InstanceClass),
		AllocatedStorage: pulumi.Int(20),

		DbName:   pulumi.String(p.DbName),
		Username: pulumi.String(p.MasterUsername),

		ManageMasterUserPassword: pulumi.Bool(true),

		DbSubnetGroupName:   subnetGroup.Name,
		VpcSecurityGroupIds: pulumi.StringArray{p.DbSgId},

		PubliclyAccessible: pulumi.Bool(false),
		MultiAz:            pulumi.Bool(p.MultiAz),

		BackupRetentionPeriod: pulumi.Int(p.BackupRetention),

		SkipFinalSnapshot:       pulumi.Bool(false),
		DeletionProtection:      pulumi.Bool(false), // TODO: if set to true, can I still have a way to delete?
		FinalSnapshotIdentifier: pulumi.String(fmt.Sprintf("%s-%s-final-snapshot", projectName, p.Environment)),

		Tags: pulumi.StringMap{
			"Environment": pulumi.String(p.Environment),
			"Name":        pulumi.String(name),
		},
	}, opts...)

	if err != nil {
		return nil, err
	}

	// masterUserSecret is only present if manageMasterUserPassword=true
	// TODO: need to figure out how to get the secret ARN
	// secretArn := db.MasterUserSecrets.Index(pulumi.Int(0)).SecretArn().ToStringPtrOutput()

	secretArn := db.MasterUserSecrets.ApplyT(func(secrets []rds.InstanceMasterUserSecret) *string {
		if len(secrets) == 0 {
			return nil
		}
		if secrets[0].SecretArn == nil {
			return nil
		}
		return secrets[0].SecretArn
	}).(pulumi.StringPtrOutput)

	return &Result{
		Instance:            db,
		Endpoint:            db.Address,
		Port:                db.Port,
		MasterUserSecretArn: secretArn,
	}, nil
}
