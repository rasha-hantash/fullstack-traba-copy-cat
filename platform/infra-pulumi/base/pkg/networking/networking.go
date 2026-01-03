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

// NewNetworking provisions a "real" VPC network layout suitable for ECS/Fargate:
//
// High-level:
//  1. Create a VPC with DNS enabled
//  2. Attach an Internet Gateway (public internet egress/ingress for public subnets) making it possible for the internet to reach resources that are in private subnets yet connected to the IGW
//  3. Discover Availability Zones and spread subnets across them (basic HA)
//  4. Split the VPC CIDR into N public + N private subnet CIDRs
//  5. Create public subnets (auto-assign public IPs) + private subnets (no public IPs)
//  6. Create a single NAT Gateway (in the first public subnet) so private subnets can reach the internet
//     for things like ECR pulls, OS package downloads, etc.
//  7. Create route tables:
//     - public RT: 0.0.0.0/0 -> IGW
//     - private RT: 0.0.0.0/0 -> NAT
//  8. Associate the appropriate RT to each subnet
//
// Returns the VPC ID and subnet IDs as outputs to be consumed by the rest of the stack.
func NewNetworking(ctx *pulumi.Context, projectName string, n Networking) (*Result, error) {
	// --- VPC ---
	// Creates the VPC "container" that all networking resources live inside.
	// DNS support/hostnames are important for ALBs, private DNS, ECS, etc.
	// This allows you to create ALBs (Application Load Balancers) to point to DNS names, i.e. app.fs0ciety.dev and api.fs0ciety.dev and have them work smoothly
	// Instead of having to use IP addresses, which are not as user friendly and easy to remember and frequently changing.
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

	// --- Internet Gateway (IGW) ---
	// Required for public subnets to have direct internet access.
	// This IGW will be used to route traffic coming in and going out of the VPC to the internet.
	// Example: A user goes to our website app.fs0ciety.dev, that traffic is redirected by the IGW to the private subent of the ECS resource
	// An internet gateway enables resources in your public subnets (such as EC2 instances) to connect to the internet
	// if the resource has a public IPv4 address or an IPv6 address.
	// Similarly, resources on the internet can initiate a connection to resources in your subnet using the
	// public IPv4 address or IPv6 address. For example, an internet gateway
	// enables you to connect to an EC2 instance in AWS using your local computer.
	// See more here: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Internet_Gateway.html
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

	// --- Availability Zones (AZs) ---
	// We distribute both public and private subnets across AZs for basic fault tolerance.
	azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{
		State: pulumi.StringRef("available"),
	}, nil)
	if err != nil {
		return nil, err
	}
	// Many production patterns assume at least 2 AZs.
	if len(azs.Names) < 2 {
		return nil, fmt.Errorf("need at least 2 AZs, found %d", len(azs.Names))
	}

	// --- CIDR planning ---
	// Split the VPC CIDR into (publicCount + privateCount) smaller CIDR blocks.
	// We then use the first N for public subnets and the rest for private subnets.
	// IP addresses enabl10.0.0.0/16 represents 65,536 IPv4 addresses from 10.0.0.0 to 10.0.255.255.
	// Example: 10.0.0.0/16 represents 65,536 IPv4 addresses from 10.0.0.0 to 10.0.255.255.
	// The first 256 addresses (10.0.0.0/24) are used for the public subnets, and the remaining 65,280 addresses (10.0.1.0/24 to 10.0.255.255/24) are used for the private subnets.
	// This is a common pattern for VPCs, and it is a good way to ensure that the subnets are evenly distributed across the AZs.
	// See more here: https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Subnets.html

	publicCidrs, privateCidrs, err := splitCidr(ctx, n.VpcCidr, n.NewBits, n.PublicCount, n.PrivateCount)
	if err != nil {
		return nil, err
	}

	// Keep references to created subnets so we can:
	// - attach NAT into a public subnet
	// - associate route tables
	// - return subnet IDs as outputs
	publicSubnets := make([]*ec2.Subnet, 0, n.PublicCount)
	privateSubnets := make([]*ec2.Subnet, 0, n.PrivateCount)

	// --- Public subnets ---
	// Public subnets:
	// - have a route to the IGW (via public route table below)
	// - map public IPs on launch for instances/tasks that need direct internet reachability
	// This is useful for the following use cases:
	// - ECS tasks that need to reach the internet
	// - ECS tasks that need to reach the internet
	// What actually lives in private subnets?
	// Always private
	// RDS
	// ECS services
	// Internal workers
	// Redis / OpenSearch / etc.
	// Usually public
	// ALB	(Application Load Balancer)
	// NAT Gateway	(Network Address Translation Gateway)
	// Bastion / SSM access instance (if you have one)	(A bastion host is a server that allows you to securely connect to other servers in your VPC.)
	for i := 0; i < n.PublicCount; i++ {
		name = fmt.Sprintf("%s-%s-public-%d", projectName, n.Environment, i+1)
		sn, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:               vpc.ID(),
			CidrBlock:           pulumi.String(publicCidrs[i]),
			AvailabilityZone:    pulumi.String(azs.Names[i%len(azs.Names)]), // round-robin which subnet goes to which AZ
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

	// --- Private subnets ---
	// Private subnets:
	// - do NOT auto-assign public IPs
	// - will route outbound internet traffic through NAT (via private route table below)
	for i := 0; i < n.PrivateCount; i++ {
		name = fmt.Sprintf("%s-%s-private-%d", projectName, n.Environment, i+1)
		sn, err := ec2.NewSubnet(ctx, name, &ec2.SubnetArgs{
			VpcId:            vpc.ID(),
			CidrBlock:        pulumi.String(privateCidrs[i]),
			AvailabilityZone: pulumi.String(azs.Names[i%len(azs.Names)]), // round-robin which subnet goes to which AZ
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

	// --- NAT Gateway ---
	// A NAT Gateway allows resources in private subnets to reach the public internet
	// WITHOUT exposing them to inbound public traffic.
	// This is useful for the following use cases:
	// - Connecting your computer to the RDS instance in the private subnet (i.e from my computer I ssm tunnel into a bastion host in the private subnet, that bastion host port forwards me to the RDS instance)
	// - Allowing an ECS backend container to send a response to an Auth0 webhook / API call
	// Note: this creates a single NAT in the first public subnet (cost-effective because NAT gateways are expensive, more expensive is one NAT per AZ).
	// More HA (and more expensive) is one NAT per AZ.
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
		SubnetId:     publicSubnets[0].ID(), // NAT lives in a public subnet
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(name),
			"Environment": pulumi.String(n.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	// --- Route tables ---
	// Public route table: default route -> IGW
	// Route tables control egress (outbound) traffic decisions. 

	// This route table enables resources in private subnets to reach the internet.
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

	// Private route table: default route -> NAT
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

	// --- Route table associations ---
	// Bind every public subnet to the public route table and every private subnet to the private route table.
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

	// --- Collect IDs for outputs ---
	// We return IDs rather than concrete resource pointers so other stacks/modules can consume them cleanly.
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

// splitCidr deterministically splits a base CIDR into smaller subnets.
//
// Example intuition:
// - base = "10.0.0.0/16"
// - newBits = 8  => creates /24 subnets
// - total = publicCount + privateCount
//
// We generate `total` CIDRs using cidrsubnet(base, newBits, netnum) for netnum in [0..total-1].
// The first `publicCount` CIDRs are returned as public subnet CIDRs; the remainder are private subnet CIDRs.
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
