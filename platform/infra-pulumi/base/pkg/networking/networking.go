package networking

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-std/sdk/go/std"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Networking struct {
	Environment  string
	VpcCidr      string
	PublicCount  int
	PrivateCount int
	NewBits      int
}

type Result struct {
	VpcId            pulumi.StringOutput
	PublicSubnetIds  pulumi.StringArrayOutput
	PrivateSubnetIds pulumi.StringArrayOutput
}

func NewNetworking(ctx *pulumi.Context, projectName string, n Networking) (*Result, error) {
	name := fmt.Sprintf("%s-%s-vpc", projectName, n.Environment)
	vpc, err := ec2.NewVpc(ctx, name, &ec2.VpcArgs{
		CidrBlock:          pulumi.String(n.VpcCidr),
		EnableDnsHostnames: pulumi.Bool(true),
		EnableDnsSupport:   pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-vpc", projectName, n.Environment)),
			"Environment": pulumi.String(n.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	name = fmt.Sprintf("%s-%s-igw", projectName, n.Environment)
	igw, err := ec2.NewInternetGateway(ctx, name, &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(fmt.Sprintf("%s-%s-igw", projectName, n.Environment)),
			"Environment": pulumi.String(n.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available"),
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(azs.Names) < 2 {
		return nil, fmt.Errorf("need at least 2 AZs, found %d", len(azs.Names))
	}

	publicCidrs, privateCidrs, err := splitCidr(ctx, n.VpcCidr, n.NewBits, n.PublicCount, n.PrivateCount)
	if err != nil {
		return nil, err
	}

	publicSubnets := make([]*ec2.Subnet, 0, n.PublicCount)
	privateSubnets := make([]*ec2.Subnet, 0, n.PrivateCount)

	for i := 0; i < n.PublicCount; i++ {
		name = fmt.Sprintf("%s-%s-public-%d", projectName, n.Environment, i+1)
		sn, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			CidrBlock:           pulumi.String(publicCidrs[i]),
			AvailabilityZone:    pulumi.String(azs.Names[i%len(azs.Names)]),
			MapPublicIpOnLaunch: pulumi.Bool(true),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String(name),
				"Environment": pulumi.String(n.Environment),
			},
		})
		if err != nil {
			return nil, err
		}
		publicSubnets = append(publicSubnets, sn)
	}

	for i := 0; i < n.PrivateCount; i++ {
		name = fmt.Sprintf("%s-%s-private-%d", projectName, n.Environment, i+1)
		sn, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(privateCidrs[i]),
			AvailabilityZone: pulumi.String(azs.Names[i%len(azs.Names)]),
			Tags: pulumi.StringMap{
				"Name":        pulumi.String(name),
				"Environment": pulumi.String(n.Environment),
			},
		})
		if err != nil {
			return nil, err
		}
		privateSubnets = append(privateSubnets, sn)
	}

	// NAT in first public subnet
	name = fmt.Sprintf("%s-%s-nat-eip", projectName, n.Environment)
	eip, err := ec2.NewEip(ctx, name, &ec2.EipArgs{
		Domain: pulumi.String("vpc"),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(n.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	name = fmt.Sprintf("%s-%s-nat-gateway", projectName, n.Environment)
	nat, err := ec2.NewNatGateway(ctx, name, &ec2.NatGatewayArgs{
		AllocationId: eip.ID(),
		SubnetId:     publicSubnets[0].ID(),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(n.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	name = fmt.Sprintf("%s-%s-public-rt", projectName, n.Environment)
	publicRt, err := ec2.NewRouteTable(ctx, name, &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(n.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	name = fmt.Sprintf("%s-%s-public-default", projectName, n.Environment)
	_, err = ec2.NewRoute(ctx, name, &ec2.RouteArgs{
		RouteTableId:         publicRt.ID(),
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		GatewayId:            igw.ID(),
	})
	if err != nil {
		return nil, err
	}

	name = fmt.Sprintf("%s-%s-private-rt", projectName, n.Environment)
	privateRt, err := ec2.NewRouteTable(ctx, name, &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(n.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	name = fmt.Sprintf("%s-%s-private-default", projectName, n.Environment)
	_, err = ec2.NewRoute(ctx, name, &ec2.RouteArgs{
		RouteTableId:         privateRt.ID(),
		DestinationCidrBlock: pulumi.String("0.0.0.0/0"),
		NatGatewayId:         nat.ID(),
	})
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(publicSubnets); i++ {
		name = fmt.Sprintf("%s-%s-public-assoc-%d", projectName, n.Environment, i+1)
		_, err = ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
			RouteTableId: publicRt.ID(),
			SubnetId:     publicSubnets[i].ID(),
		})
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(privateSubnets); i++ {
		name = fmt.Sprintf("%s-%s-private-assoc-%d", projectName, n.Environment, i+1)
		_, err = ec2.NewRouteTableAssociation(ctx, name, &ec2.RouteTableAssociationArgs{
			RouteTableId: privateRt.ID(),
			SubnetId:     privateSubnets[i].ID(),
		})
		if err != nil {
			return nil, err
		}
	}

	pubIds := pulumi.StringArray{}
	for _, s := range publicSubnets {
		pubIds = append(pubIds, s.ID())
	}
	privIds := pulumi.StringArray{}
	for _, s := range privateSubnets {
		privIds = append(privIds, s.ID())
	}

	return &Result{
		VpcId:            vpc.ID().ToStringOutput(),
		PublicSubnetIds:  pubIds.ToStringArrayOutput(),
		PrivateSubnetIds: privIds.ToStringArrayOutput(),
	}, nil
}

// returns []string CIDRs
func splitCidr(ctx *pulumi.Context, base string, newBits, publicCount, privateCount int) ([]string, []string, error) {
	total := publicCount + privateCount
	all := make([]string, 0, total)

	for netnum := 0; netnum < total; netnum++ {
		r, err := std.Cidrsubnet(ctx, &std.CidrsubnetArgs{
			Input:   base,
			Newbits: newBits,
			Netnum:  netnum,
		})
		if err != nil {
			return nil, nil, err
		}
		all = append(all, r.Result)
	}

	return all[:publicCount], all[publicCount:], nil
}
