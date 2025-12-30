package ecr

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EcrRepository struct {
	Name        string
	Environment string
}

func NewEcrRepository(ctx *pulumi.Context, config EcrRepository) (*ecr.Repository, error) {
	repo, err := ecr.NewRepository(ctx, config.Name+"-"+config.Environment, &ecr.RepositoryArgs{
		Name: pulumi.String(config.Name + "-" + config.Environment),
		ImageScanningConfiguration: &ecr.RepositoryImageScanningConfigurationArgs{
			ScanOnPush: pulumi.Bool(true),
		},
		ImageTagMutability: pulumi.String("MUTABLE"),
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}
