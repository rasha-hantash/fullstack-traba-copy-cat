package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type baseOutputs struct {
	// ECS
	ecsClusterArn    pulumi.StringOutput
	executionRoleArn pulumi.StringOutput
	taskRoleArn      pulumi.StringOutput

	// Networking for awsvpc
	privateSubnetIds pulumi.StringArrayOutput

	// Security groups (service SGs, not ALB SGs)
	frontendSgId pulumi.StringOutput
	backendSgId  pulumi.StringOutput

	// Target groups created in base ALB module
	frontendTargetGroupArn pulumi.StringOutput
	backendTargetGroupArn  pulumi.StringOutput

	// ECR repository URL
	ecrRepoUrl pulumi.StringOutput
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		fmt.Println("cfg", cfg)

		awsRegion := cfg.Require("awsRegion")
		env := ctx.Stack() // "prod" for prod stack, "staging" for staging, etc.

		// === Base stack reference ===
		baseStack := cfg.Require("baseStack")
		baseRef, err := pulumi.NewStackReference(ctx, baseStack, nil)
		if err != nil {
			return err
		}
		base := loadBaseOutputs(baseRef)

		// ecrRepoUrl := baseRef.GetOutput(pulumi.String("ecrRepoUrl"))



		// ✅ These should match what your Taskfile sets:
		// pulumi config set frontendImage <full_ecr_uri:tag>
		// pulumi config set backendImage <full_ecr_uri:tag>
		frontendImage := pulumi.Sprintf("%s:%s", base.ecrRepoUrl, cfg.Require("frontendImageTag"))
		backendImage := pulumi.Sprintf("%s:%s", base.ecrRepoUrl, cfg.Require("backendImageTag"))

		// Log groups
		feLogs, err := cloudwatch.NewLogGroup(ctx, "frontend-logs", &cloudwatch.LogGroupArgs{
			Name:            pulumi.String(fmt.Sprintf("/ecs/%s-frontend", ctx.Stack())),
			RetentionInDays: pulumi.Int(30),
		})
		if err != nil {
			return err
		}

		beLogs, err := cloudwatch.NewLogGroup(ctx, "backend-logs", &cloudwatch.LogGroupArgs{
			Name:            pulumi.String(fmt.Sprintf("/ecs/%s-backend", ctx.Stack())),
			RetentionInDays: pulumi.Int(30),
		})
		if err != nil {
			return err
		}

		// === Task Definitions ===
		feContainerDefs := pulumi.Sprintf(`[
			{
			  "name": "frontend",
			  "image": %q,
			  "portMappings": [{"containerPort": 80, "protocol": "tcp"}],
			  "essential": true,
			  "environment": [
				{"name": "ENV", "value": %q},
				{"name": "AWS_REGION", "value": %q}
			  ],
			  "logConfiguration": {
				"logDriver": "awslogs",
				"options": {
				  "awslogs-group": %q,
				  "awslogs-region": %q,
				  "awslogs-stream-prefix": "ecs"
				}
			  }
			}
		]`, frontendImage, env, awsRegion, feLogs.Name, awsRegion)

		feTaskDef, err := ecs.NewTaskDefinition(ctx, "frontend-taskdef", &ecs.TaskDefinitionArgs{
			Family:                  pulumi.String(fmt.Sprintf("%s-frontend", ctx.Stack())),
			NetworkMode:             pulumi.String("awsvpc"),
			RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
			Cpu:                     pulumi.String("256"),
			Memory:                  pulumi.String("512"),
			ExecutionRoleArn:        base.executionRoleArn,
			TaskRoleArn:             base.taskRoleArn,
			ContainerDefinitions:    feContainerDefs,
		})
		if err != nil {
			return err
		}

		beContainerDefs := pulumi.Sprintf(`[
			{
				"name": "backend",
				"image": %q,
				"portMappings": [{"containerPort": 3000, "protocol": "tcp"}],
				"essential": true,
				"environment": [
					{"name": "ENV", "value": %q},
					{"name": "AWS_REGION", "value": %q}
				],
				"logConfiguration": {
					"logDriver": "awslogs",
					"options": {
						"awslogs-group": %q,
						"awslogs-region": %q,
						"awslogs-stream-prefix": "ecs"
					}
				}
			}
		]`, backendImage, env, awsRegion, beLogs.Name, awsRegion)

		beTaskDef, err := ecs.NewTaskDefinition(ctx, "backend-taskdef", &ecs.TaskDefinitionArgs{
			Family:                  pulumi.String(fmt.Sprintf("%s-backend", ctx.Stack())),
			NetworkMode:             pulumi.String("awsvpc"),
			RequiresCompatibilities: pulumi.StringArray{pulumi.String("FARGATE")},
			Cpu:                     pulumi.String("256"),
			Memory:                  pulumi.String("512"),
			ExecutionRoleArn:        base.executionRoleArn,
			TaskRoleArn:             base.taskRoleArn,
			ContainerDefinitions:    beContainerDefs,
		})
		if err != nil {
			return err
		}

		// === ECS Services ===
		feSvc, err := ecs.NewService(ctx, "frontend-service", &ecs.ServiceArgs{
			Name:           pulumi.String(fmt.Sprintf("%s-frontend", ctx.Stack())),
			Cluster:        base.ecsClusterArn,
			TaskDefinition: feTaskDef.Arn,
			DesiredCount:   pulumi.Int(1),
			LaunchType:     pulumi.String("FARGATE"),

			NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
				AssignPublicIp: pulumi.Bool(false),
				Subnets:        base.privateSubnetIds,
				SecurityGroups: pulumi.StringArray{base.frontendSgId},
			},

			LoadBalancers: ecs.ServiceLoadBalancerArray{
				&ecs.ServiceLoadBalancerArgs{
					TargetGroupArn: base.frontendTargetGroupArn,
					ContainerName:  pulumi.String("frontend"),
					ContainerPort:  pulumi.Int(80),
				},
			},

			DeploymentMinimumHealthyPercent: pulumi.Int(50),
			DeploymentMaximumPercent:        pulumi.Int(200),
		})
		if err != nil {
			return err
		}

		beSvc, err := ecs.NewService(ctx, "backend-service", &ecs.ServiceArgs{
			Name:           pulumi.String(fmt.Sprintf("%s-backend", ctx.Stack())),
			Cluster:        base.ecsClusterArn,
			TaskDefinition: beTaskDef.Arn,
			DesiredCount:   pulumi.Int(1),
			LaunchType:     pulumi.String("FARGATE"),

			NetworkConfiguration: &ecs.ServiceNetworkConfigurationArgs{
				AssignPublicIp: pulumi.Bool(false),
				Subnets:        base.privateSubnetIds,
				SecurityGroups: pulumi.StringArray{base.backendSgId},
			},

			LoadBalancers: ecs.ServiceLoadBalancerArray{
				&ecs.ServiceLoadBalancerArgs{
					TargetGroupArn: base.backendTargetGroupArn,
					ContainerName:  pulumi.String("backend"),
					ContainerPort:  pulumi.Int(3000),
				},
			},

			DeploymentMinimumHealthyPercent: pulumi.Int(50),
			DeploymentMaximumPercent:        pulumi.Int(200),
		})
		if err != nil {
			return err
		}

		// Helpful exports
		ctx.Export("frontendImage", frontendImage)
		ctx.Export("backendImage", backendImage)
		ctx.Export("frontendServiceName", feSvc.Name)
		ctx.Export("backendServiceName", beSvc.Name)

		return nil
	})
}

func loadBaseOutputs(ref *pulumi.StackReference) baseOutputs {
	return baseOutputs{
		ecsClusterArn:          requireString(ref, "ecsClusterArn"),
		executionRoleArn:       requireString(ref, "ecsExecutionRoleArn"),
		taskRoleArn:            requireString(ref, "ecsTaskRoleArn"),
		privateSubnetIds:       requireStringArray(ref, "privateSubnetIds"),
		frontendSgId:           requireString(ref, "frontendSgId"),
		backendSgId:            requireString(ref, "backendSgId"),
		frontendTargetGroupArn: requireString(ref, "frontendTargetGroupArn"),
		backendTargetGroupArn:  requireString(ref, "backendTargetGroupArn"),
		ecrRepoUrl:            requireString(ref, "ecrRepoUrl"),
	}
}

func requireString(ref *pulumi.StackReference, key string) pulumi.StringOutput {
	return ref.GetOutput(pulumi.String(key)).ApplyT(func(v interface{}) (string, error) {
		if v == nil {
			return "", fmt.Errorf("missing required base output: %s", key)
		}
		s, ok := v.(string)
		if !ok || s == "" {
			return "", fmt.Errorf("base output %s is not a non-empty string", key)
		}
		return s, nil
	}).(pulumi.StringOutput)
}

func requireStringArray(ref *pulumi.StackReference, key string) pulumi.StringArrayOutput {
	return ref.GetOutput(pulumi.String(key)).ApplyT(func(v interface{}) ([]string, error) {
		if v == nil {
			return nil, fmt.Errorf("missing required base output: %s", key)
		}
		raw, ok := v.([]interface{})
		if !ok {
			if s, ok2 := v.([]string); ok2 {
				return s, nil
			}
			return nil, fmt.Errorf("base output %s is not an array", key)
		}
		out := make([]string, 0, len(raw))
		for _, item := range raw {
			str, ok := item.(string)
			if !ok || str == "" {
				return nil, fmt.Errorf("base output %s contains non-string/empty item", key)
			}
			out = append(out, str)
		}
		return out, nil
	}).(pulumi.StringArrayOutput)
}
