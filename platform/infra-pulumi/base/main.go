package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/config"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/database"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/dns"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/ecr"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/ecscluster"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/loadbalancer"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/networking"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/secrets"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/security"
	"github.com/rasha-hantash/fullstack-traba-copy-cat/platform/infra-pulumi/base/pkg/ssmaccess"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.Must(ctx)

		net, err := networking.NewNetworking(ctx, cfg.ProjectName, networking.Networking{
			Environment:  cfg.Environment,
			VpcCidr:      cfg.VpcCidr,
			PublicCount:  2,
			PrivateCount: 2,
			NewBits:      7, // matches your old terraform cidrsubnet(newbits=7)
		})
		if err != nil {
			return err
		}

		sgs, err := security.NewSecurityGroups(ctx, cfg.ProjectName, security.SecurityGroups{
			Environment:           cfg.Environment,
			VpcId:                 net.VpcId,
			FrontendContainerPort: 80,
			BackendContainerPort:  3000,
			DbContainerPort:       5432,
		})
		if err != nil {
			return err
		}

		cluster, err := ecscluster.NewEcsCluster(ctx, cfg.ProjectName, ecscluster.EcsCluster{
			Environment: cfg.Environment,
		})
		if err != nil {
			return err
		}

		iamRes, err := ecscluster.NewTaskIam(ctx, ecscluster.TaskIam{
			Environment: cfg.Environment,
			ProjectName: cfg.ProjectName,
			SecretEnv:   "prod", // 👈 if this stack is prod, keep as cfg.Environment; if you *always* want prod secrets, set "prod"
			Service:     "core",
		})
		if err != nil {
			return err
		}

		/**
		This is only doing ACM certificate + DNS validation:

			1. Find the hosted zone (Route53 container) for fs0ciety.dev (hosted zone name)

			2. Request an ACM certificate for the domain name (environment.hostedZoneName)

			3. ACM says: “prove you own this domain by creating a DNS record”

			4. You create that validation record in Route53

			5. You create CertificateValidation so Pulumi waits until ACM says “cert issued”
		**/

		d, err := dns.NewDnsRecords(ctx, cfg.ProjectName, dns.DnsRecords{
			Environment:     cfg.Environment,
			HostedZoneName:  cfg.HostedZoneName,
			IncludeWildcard: true, // optional: if you want *.fs0ciety.dev too
		})
		if err != nil {
			return err
		}

		frontendDomain := "app." + cfg.HostedZoneName
		backendDomain := "api." + cfg.HostedZoneName
		// Update your ALB calls to use the specific security groups
		edgeAlb, err := loadbalancer.NewEdgeAlb(ctx, cfg.ProjectName, loadbalancer.EdgeAlb{
			Environment:     cfg.Environment,
			VpcId:           net.VpcId,
			PublicSubnetIds: net.PublicSubnetIds,
			AlbSgId:         sgs.AlbSgId,
			CertArn:         d.CertArn,

			Frontend: loadbalancer.HostRule{
				Hostnames:       []string{frontendDomain},
				TargetPort:      80,
				HealthCheckPath: "/health",
				Priority:        10,
			},
			Backend: loadbalancer.HostRule{
				Hostnames:       []string{backendDomain},
				TargetPort:      3000,
				HealthCheckPath: "/health",
				Priority:        20,
			},
		})
		if err != nil {
			return err
		}

		// app.example.com -> ALB
		_, err = dns.NewAliasARecord(ctx, cfg.ProjectName, dns.AliasARecord{
			Environment:      cfg.Environment,
			Service:          "frontend",
			Route53ZoneId:    d.HostedZoneId,
			RecordDomainName: frontendDomain,
			AlbZoneId:        edgeAlb.AlbZoneId,
			AlbDnsName:       edgeAlb.AlbDnsName,
		})

		if err != nil {
			return err
		}

		_, err = dns.NewAliasARecord(ctx, cfg.ProjectName, dns.AliasARecord{
			Environment:      cfg.Environment,
			Service:          "backend",
			Route53ZoneId:    d.HostedZoneId,
			RecordDomainName: backendDomain,
			AlbZoneId:        edgeAlb.AlbZoneId,
			AlbDnsName:       edgeAlb.AlbDnsName,
		})

		if err != nil {
			return err
		}

		db, err := database.NewPostgres(ctx, cfg.ProjectName, database.Postgres{
			Environment:      cfg.Environment,
			PrivateSubnetIds: net.PrivateSubnetIds,
			DbSgId:           sgs.DbSgId,

			DbName:         cfg.Postgres.DbName,
			MasterUsername: cfg.Postgres.DbUser,

			InstanceClass:   cfg.Postgres.DatabaseInstanceType,
			MultiAz:         cfg.Postgres.DatabaseMultiAz,
			BackupRetention: cfg.Postgres.DatabaseBackupRetentionDays,
		})
		if err != nil {
			return err
		}

		// --- SSM-only DB access instance (for port forwarding / operator access) ---
		ssm, err := ssmaccess.NewDbSsmAccess(ctx, cfg.ProjectName, ssmaccess.DbSsmAccess{
			Environment: cfg.Environment,
			VpcId:       net.VpcId,

			// pick first private subnet
			SubnetId: net.PrivateSubnetIds.Index(pulumi.Int(0)),

			DbSgId: sgs.DbSgId,

			InstanceType: "t3.nano",
		})
		if err != nil {
			return err
		}

		// allow db-access (ssm) instance to reach rds on 5432
		_, err = ec2.NewSecurityGroupRule(ctx, "db-allow-ssm-access", &ec2.SecurityGroupRuleArgs{
			Type:                  pulumi.String("ingress"),
			SecurityGroupId:       sgs.DbSgId, // the DB SG
			FromPort:              pulumi.Int(5432),
			ToPort:                pulumi.Int(5432),
			Protocol:              pulumi.String("tcp"),
			SourceSecurityGroupId: ssm.SecurityGroupId, // the SSM instance SG
			Description:           pulumi.String("Allow Postgres from SSM db-access instance"),
		})
		if err != nil {
			return err
		}

		// we will use this in order to store the DB credentials in AWS Secrets Manager that will be initually
		// created when ManageMasterUserPassword=true, RDS creates a Secrets Manager secret
		// and exposes its ARN here.
		awsSecrets, err := secrets.NewAwsSecrets(
			ctx,
			cfg.ProjectName,
			secrets.AwsSecrets{
				Environment: cfg.Environment,
				Service:     "core",
			})
		if err != nil {
			return err
		}

		localAwsSecrets, err := secrets.NewAwsSecrets(ctx, cfg.ProjectName, secrets.AwsSecrets{
			Environment: "local",
			Service:     "core",
		})
		if err != nil {
			return err
		}

		repo, err := ecr.NewEcrRepository(ctx, ecr.EcrRepository{
			Name:        cfg.ProjectName,
			Environment: cfg.Environment,
		})
		if err != nil {
			return err
		}

		// ---- Exports = base -> services contract ----

		// 1) Networking
		ctx.Export("vpcId", net.VpcId)
		ctx.Export("publicSubnetIds", net.PublicSubnetIds)
		ctx.Export("privateSubnetIds", net.PrivateSubnetIds)

		// 2) DNS + domains (what hostnames exist + where the zone is)
		ctx.Export("hostedZoneId", d.HostedZoneId)
		ctx.Export("certArn", d.CertArn)
		ctx.Export("frontendDomain", pulumi.String(frontendDomain))
		ctx.Export("backendDomain", pulumi.String(backendDomain))

		// 3) Edge ALB + routing primitives (what to attach services to)
		ctx.Export("albArn", edgeAlb.AlbArn)
		ctx.Export("albDnsName", edgeAlb.AlbDnsName)
		ctx.Export("albZoneId", edgeAlb.AlbZoneId)
		ctx.Export("httpsListenerArn", edgeAlb.HttpsListenerArn)

		ctx.Export("frontendTargetGroupArn", edgeAlb.FrontendTargetGroupArn)
		ctx.Export("backendTargetGroupArn", edgeAlb.BackendTargetGroupArn)

		ctx.Export("frontendListenerRuleArn", edgeAlb.FrontendRuleArn)
		ctx.Export("backendListenerRuleArn", edgeAlb.BackendRuleArn)

		// 4) ECS + security groups (what services will run in + what SGs to use)
		ctx.Export("ecsClusterArn", cluster.ClusterArn)
		ctx.Export("ecsExecutionRoleArn", iamRes.ExecutionRoleArn)
		ctx.Export("ecsTaskRoleArn", iamRes.TaskRoleArn)

		ctx.Export("albSgId", sgs.AlbSgId)
		ctx.Export("frontendSgId", sgs.FrontendSgId)
		ctx.Export("backendSgId", sgs.BackendSgId)

		// 5) Database
		ctx.Export("dbEndpoint", db.Endpoint)
		ctx.Export("dbPort", db.Port)
		ctx.Export("dbMasterUserSecretArn", db.MasterUserSecretArn)

		// 6) General Secrets Manager container(s)
		ctx.Export("awsSecretsSecretName", awsSecrets.Name)
		ctx.Export("awsSecretsSecretArn", awsSecrets.Arn)
		ctx.Export("localAwsSecretsSecretName", localAwsSecrets.Name)
		ctx.Export("localAwsSecretsSecretArn", localAwsSecrets.Arn)
		ctx.Export("ecrRepoUrl", repo.RepositoryUrl)

		// 7) SSM-only DB access instance
		ctx.Export("ssmDbAccessInstanceId", ssm.InstanceId)
		ctx.Export("ssmDbAccessSgId", ssm.SecurityGroupId)

		return nil
	})
}
