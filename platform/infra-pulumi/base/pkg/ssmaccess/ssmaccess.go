package ssmaccess

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DbSsmAccess struct {
	Environment string

	VpcId pulumi.StringInput

	// put this instance in a private subnet (recommended)
	SubnetId pulumi.StringInput

	// allow egress to DB + SSM (443)
	DbSgId pulumi.StringInput

	// sizing / instance opts
	InstanceType string // e.g. "t3.nano"
}

type Result struct {
	InstanceId      pulumi.StringOutput
	SecurityGroupId pulumi.StringOutput
}

func NewDbSsmAccess(ctx *pulumi.Context, projectName string, a DbSsmAccess, opts ...pulumi.ResourceOption) (*Result, error) {
	namePrefix := fmt.Sprintf("%s-%s-db-ssm", projectName, a.Environment)

	// -----------------------------
	// IAM role + instance profile
	// -----------------------------
	assumeRolePolicy := `{
	  "Version": "2012-10-17",
	  "Statement": [{
	    "Effect": "Allow",
	    "Principal": { "Service": "ec2.amazonaws.com" },
	    "Action": "sts:AssumeRole"
	  }]
	}`

	role, err := iam.NewRole(ctx, namePrefix+"-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(assumeRolePolicy),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(namePrefix + "-role"),
			"Environment": pulumi.String(a.Environment),
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	// This is the *only* policy required for SSM managed instances.
	_, err = iam.NewRolePolicyAttachment(ctx, namePrefix+"-ssm-attach", &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"),
	}, opts...)
	if err != nil {
		return nil, err
	}

	profile, err := iam.NewInstanceProfile(ctx, namePrefix+"-profile", &iam.InstanceProfileArgs{
		Role: role.Name,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(namePrefix + "-profile"),
			"Environment": pulumi.String(a.Environment),
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	// -----------------------------
	// Security group: SSM-only, no inbound
	// -----------------------------
	sg, err := ec2.NewSecurityGroup(ctx, namePrefix+"-sg", &ec2.SecurityGroupArgs{
		VpcId:       a.VpcId.ToStringOutput(),
		Description: pulumi.String("SSM-only DB access instance (no inbound)"),
		Egress: ec2.SecurityGroupEgressArray{
			// allow HTTPS out for SSM
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(443),
				ToPort:     pulumi.Int(443),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
			// allow postgres out to the DB SG
			&ec2.SecurityGroupEgressArgs{
				Protocol: pulumi.String("tcp"),
				FromPort: pulumi.Int(5432),
				ToPort:   pulumi.Int(5432),
				SecurityGroups: pulumi.StringArray{
					a.DbSgId.ToStringOutput(),
				},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(namePrefix + "-sg"),
			"Environment": pulumi.String(a.Environment),
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	// -----------------------------
	// AMI: Amazon Linux 2023
	// -----------------------------
	ami, err := ec2.LookupAmi(ctx, &ec2.LookupAmiArgs{
		MostRecent: pulumi.BoolRef(true),
		Owners:     []string{"amazon"},
		Filters: []ec2.GetAmiFilter{
			{
				Name:   "name",
				Values: []string{"al2023-ami-*-x86_64"},
			},
			{
				Name:   "state",
				Values: []string{"available"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	itype := a.InstanceType
	if itype == "" {
		itype = "t3.nano"
	}

	// -----------------------------
	// EC2 instance (no keypair, no SSH)
	// -----------------------------
	inst, err := ec2.NewInstance(ctx, namePrefix, &ec2.InstanceArgs{
		Ami:                      pulumi.String(ami.Id),
		InstanceType:             pulumi.String(itype),
		SubnetId:                 a.SubnetId.ToStringOutput(),
		IamInstanceProfile:       profile.Name,
		VpcSecurityGroupIds:      pulumi.StringArray{sg.ID().ToStringOutput()},
		AssociatePublicIpAddress: pulumi.Bool(false), // keep it private
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(namePrefix),
			"Environment": pulumi.String(a.Environment),
			"Role":        pulumi.String("db-ssm-access"),
		},
	}, opts...)
	if err != nil {
		return nil, err
	}

	return &Result{
		InstanceId:      inst.ID().ToStringOutput(),
		SecurityGroupId: sg.ID().ToStringOutput(),
	}, nil
}
