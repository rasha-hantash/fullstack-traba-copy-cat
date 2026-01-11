package security

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SecurityGroups struct {
	Environment           string
	VpcId                 pulumi.StringInput
	FrontendContainerPort int
	BackendContainerPort  int
	DbContainerPort       int
}

type Result struct {
	// ALB Security Group (single)
	AlbSgId pulumi.StringOutput

	// ECS Service Security Groups
	FrontendSgId pulumi.StringOutput
	BackendSgId  pulumi.StringOutput

	// Database Security Group
	DbSgId pulumi.StringOutput
}

func NewSecurityGroups(ctx *pulumi.Context, projectName string, sgp SecurityGroups) (*Result, error) {
	name := fmt.Sprintf("%s-%s-alb-sg", projectName, sgp.Environment)
	albSg, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		VpcId:       sgp.VpcId,
		Description: pulumi.String("Edge ALB security group"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(80),
				ToPort:      pulumi.Int(80),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("Allow HTTP from anywhere"),
			},
			&ec2.SecurityGroupIngressArgs{
				Protocol:    pulumi.String("tcp"),
				FromPort:    pulumi.Int(443),
				ToPort:      pulumi.Int(443),
				CidrBlocks:  pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				Description: pulumi.String("Allow HTTPS from anywhere"),
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(sgp.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	// Frontend ECS Service Security Group
	name = fmt.Sprintf("%s-%s-frontend-sg", projectName, sgp.Environment)
	frontendSg, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		VpcId:       sgp.VpcId,
		Description: pulumi.String("Frontend ECS service security group"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(sgp.FrontendContainerPort),
				ToPort:         pulumi.Int(sgp.FrontendContainerPort),
				SecurityGroups: pulumi.StringArray{albSg.ID().ToStringOutput()},
				Description:    pulumi.String("Allow traffic from frontend ALB"),
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(sgp.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	// Backend ECS Service Security Group
	name = fmt.Sprintf("%s-%s-backend-sg", projectName, sgp.Environment)
	backendSg, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		VpcId:       sgp.VpcId,
		Description: pulumi.String("Backend ECS service security group"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(sgp.BackendContainerPort),
				ToPort:         pulumi.Int(sgp.BackendContainerPort),
				SecurityGroups: pulumi.StringArray{albSg.ID().ToStringOutput()},
				Description:    pulumi.String("Allow traffic from backend ALB"),
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(sgp.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	// Database Security Group
	name = fmt.Sprintf("%s-%s-db-sg", projectName, sgp.Environment)
	dbSg, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		VpcId:       sgp.VpcId,
		Description: pulumi.String("RDS Postgres security group"),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:       pulumi.String("tcp"),
				FromPort:       pulumi.Int(sgp.DbContainerPort),
				ToPort:         pulumi.Int(sgp.DbContainerPort),
				SecurityGroups: pulumi.StringArray{backendSg.ID().ToStringOutput()},
				Description:    pulumi.String("Allow Postgres from backend service"),
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(sgp.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	return &Result{
		AlbSgId:      albSg.ID().ToStringOutput(),
		FrontendSgId: frontendSg.ID().ToStringOutput(),
		BackendSgId:  backendSg.ID().ToStringOutput(),
		DbSgId:       dbSg.ID().ToStringOutput(),
	}, nil
}
