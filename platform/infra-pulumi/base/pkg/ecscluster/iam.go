package ecscluster

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TaskIam struct {
	Environment string

	// if you want “tight” secret access:
	ProjectName string // "fs0ciety"
	SecretEnv   string // usually same as Environment; "prod" in your case
	Service     string // "core"
}

type IamResult struct {
	ExecutionRoleArn pulumi.StringOutput
	TaskRoleArn      pulumi.StringOutput
}

func NewTaskIam(ctx *pulumi.Context, t TaskIam) (*IamResult, error) {
	// --- AssumeRolePolicy for ECS tasks ---
	assumeRolePolicy, err := json.Marshal(map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect": "Allow",
				"Principal": map[string]any{
					"Service": "ecs-tasks.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// --- Execution Role (pull image from ECR + write logs) ---
	execRoleName := fmt.Sprintf("%s-%s-ecs-execution-role", t.ProjectName, t.Environment)
	execRole, err := iam.NewRole(ctx, execRoleName, &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(string(assumeRolePolicy)),
		Tags: pulumi.StringMap{
			"Environment": pulumi.String(t.Environment),
			"Name":        pulumi.String(execRoleName),
		},
	})
	if err != nil {
		return nil, err
	}

	// Attach AWS managed execution policy (recommended)
	_, err = iam.NewRolePolicyAttachment(ctx, execRoleName+"-attach", &iam.RolePolicyAttachmentArgs{
		Role:      execRole.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"),
	})
	if err != nil {
		return nil, err
	}

	// --- Task Role (your app permissions go here; Secrets Manager read, etc.) ---
	taskRoleName := fmt.Sprintf("%s-%s-ecs-task-role", t.ProjectName, t.Environment)
	taskRole, err := iam.NewRole(ctx, taskRoleName, &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(string(assumeRolePolicy)),
		Tags: pulumi.StringMap{
			"Environment": pulumi.String(t.Environment),
			"Name":        pulumi.String(taskRoleName),
		},
	})
	if err != nil {
		return nil, err
	}

	// --- Tight Secrets Manager policy (only fs0ciety-prod-core-config*) ---
	caller, err := aws.GetCallerIdentity(ctx, nil, nil)
	if err != nil {
		return nil, err
	}
	region, err := aws.GetRegion(ctx, nil, nil)
	if err != nil {
		return nil, err
	}

	secretName := fmt.Sprintf("%s-%s-%s-config", t.ProjectName, t.SecretEnv, t.Service) // fs0ciety-prod-core-config
	secretArn := fmt.Sprintf(
		"arn:aws:secretsmanager:%s:%s:secret:%s*",
		region.Name,
		caller.AccountId,
		secretName,
	)

	policyDoc := pulumi.String(fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"secretsmanager:GetSecretValue",
					"secretsmanager:DescribeSecret"
				],
				"Resource": "%s"
			}
		]
	}`, secretArn))

	_, err = iam.NewRolePolicy(ctx, taskRoleName+"-secrets-read", &iam.RolePolicyArgs{
		Role:   taskRole.Name,
		Policy: policyDoc,
	})
	if err != nil {
		return nil, err
	}

	return &IamResult{
		ExecutionRoleArn: execRole.Arn,
		TaskRoleArn:      taskRole.Arn,
	}, nil
}
