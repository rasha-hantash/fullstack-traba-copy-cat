package ecscluster

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EcsCluster struct {
	Environment string
}

type Result struct {
	ClusterArn pulumi.StringOutput
}

func NewEcsCluster(ctx *pulumi.Context, projectName string, e EcsCluster) (*Result, error) {
	name := fmt.Sprintf("%s-%s-cluster", projectName, e.Environment)
	cluster, err := ecs.NewCluster(ctx, name, &ecs.ClusterArgs{
		Name: pulumi.String(name),
		Settings: ecs.ClusterSettingArray{
			&ecs.ClusterSettingArgs{Name: pulumi.String("containerInsights"), Value: pulumi.String("enabled")},
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(e.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	return &Result{
		ClusterArn: cluster.Arn,
	}, nil
}
