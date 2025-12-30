package loadbalancer

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type HostRule struct {
	Hostnames       []string
	TargetPort      int
	HealthCheckPath string
	Priority        int
}

type EdgeAlb struct {
	Environment     string
	VpcId           pulumi.StringInput
	PublicSubnetIds pulumi.StringArrayInput
	AlbSgId         pulumi.StringInput
	CertArn         pulumi.StringInput

	Frontend HostRule
	Backend  HostRule
}

type Result struct {
	AlbArn     pulumi.StringOutput
	AlbDnsName pulumi.StringOutput
	AlbZoneId  pulumi.StringOutput

	HttpsListenerArn pulumi.StringOutput

	FrontendTargetGroupArn pulumi.StringOutput
	BackendTargetGroupArn  pulumi.StringOutput

	FrontendRuleArn pulumi.StringOutput
	BackendRuleArn  pulumi.StringOutput
}

func NewEdgeAlb(ctx *pulumi.Context, projectName string, e EdgeAlb) (*Result, error) {
	albName := fmt.Sprintf("%s-%s-edge-alb", projectName, e.Environment)

	albRes, err := lb.NewLoadBalancer(ctx, albName, &lb.LoadBalancerArgs{
		LoadBalancerType: pulumi.String("application"),
		SecurityGroups:   pulumi.StringArray{e.AlbSgId},
		Subnets:          e.PublicSubnetIds,
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(albName),
			"Environment": pulumi.String(e.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	// --- target groups ---
	feTgName := fmt.Sprintf("%s-%s-fe-tg", projectName, e.Environment)
	feTg, err := lb.NewTargetGroup(ctx, feTgName, &lb.TargetGroupArgs{
		TargetType: pulumi.String("ip"),
		Port:       pulumi.Int(e.Frontend.TargetPort),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      e.VpcId,
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Path:               pulumi.String(e.Frontend.HealthCheckPath),
			Matcher:            pulumi.String("200-399"),
			Interval:           pulumi.Int(30),
			Timeout:            pulumi.Int(5),
			HealthyThreshold:   pulumi.Int(2),
			UnhealthyThreshold: pulumi.Int(2),
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(feTgName),
			"Environment": pulumi.String(e.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	beTgName := fmt.Sprintf("%s-%s-be-tg", projectName, e.Environment)
	beTg, err := lb.NewTargetGroup(ctx, beTgName, &lb.TargetGroupArgs{
		TargetType: pulumi.String("ip"),
		Port:       pulumi.Int(e.Backend.TargetPort),
		Protocol:   pulumi.String("HTTP"),
		VpcId:      e.VpcId,
		HealthCheck: &lb.TargetGroupHealthCheckArgs{
			Path:               pulumi.String(e.Backend.HealthCheckPath),
			Matcher:            pulumi.String("200-399"),
			Interval:           pulumi.Int(30),
			Timeout:            pulumi.Int(5),
			HealthyThreshold:   pulumi.Int(2),
			UnhealthyThreshold: pulumi.Int(2),
		},
		Tags: pulumi.StringMap{
			"Name":        pulumi.String(beTgName),
			"Environment": pulumi.String(e.Environment),
		},
	})
	if err != nil {
		return nil, err
	}

	// --- listener 80: redirect ---
	httpName := fmt.Sprintf("%s-%s-edge-http-80", projectName, e.Environment)
	_, err = lb.NewListener(ctx, httpName, &lb.ListenerArgs{
		LoadBalancerArn: albRes.Arn,
		Port:            pulumi.Int(80),
		Protocol:        pulumi.String("HTTP"),
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type: pulumi.String("redirect"),
				Redirect: &lb.ListenerDefaultActionRedirectArgs{
					Port:       pulumi.String("443"),
					Protocol:   pulumi.String("HTTPS"),
					StatusCode: pulumi.String("HTTP_301"),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// --- listener 443: default fixed response (nice base-only behavior) ---
	httpsName := fmt.Sprintf("%s-%s-edge-https-443", projectName, e.Environment)
	httpsListener, err := lb.NewListener(ctx, httpsName, &lb.ListenerArgs{
		LoadBalancerArn: albRes.Arn,
		Port:            pulumi.Int(443),
		Protocol:        pulumi.String("HTTPS"),
		SslPolicy:       pulumi.String("ELBSecurityPolicy-TLS13-1-2-2021-06"),
		CertificateArn:  e.CertArn,
		DefaultActions: lb.ListenerDefaultActionArray{
			&lb.ListenerDefaultActionArgs{
				Type: pulumi.String("fixed-response"),
				FixedResponse: &lb.ListenerDefaultActionFixedResponseArgs{
					ContentType: pulumi.String("text/plain"),
					MessageBody: pulumi.String("base deployed ✅ (services not deployed yet)"),
					StatusCode:  pulumi.String("200"),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// --- host-based rules ---
	feRuleName := fmt.Sprintf("%s-%s-edge-rule-frontend", projectName, e.Environment)
	feRule, err := lb.NewListenerRule(ctx, feRuleName, &lb.ListenerRuleArgs{
		ListenerArn: httpsListener.Arn,
		Priority:    pulumi.Int(e.Frontend.Priority),
		Actions: lb.ListenerRuleActionArray{
			&lb.ListenerRuleActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: feTg.Arn,
			},
		},
		Conditions: lb.ListenerRuleConditionArray{
			&lb.ListenerRuleConditionArgs{
				HostHeader: &lb.ListenerRuleConditionHostHeaderArgs{
					Values: pulumi.ToStringArray(e.Frontend.Hostnames),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	beRuleName := fmt.Sprintf("%s-%s-edge-rule-backend", projectName, e.Environment)
	beRule, err := lb.NewListenerRule(ctx, beRuleName, &lb.ListenerRuleArgs{
		ListenerArn: httpsListener.Arn,
		Priority:    pulumi.Int(e.Backend.Priority),
		Actions: lb.ListenerRuleActionArray{
			&lb.ListenerRuleActionArgs{
				Type:           pulumi.String("forward"),
				TargetGroupArn: beTg.Arn,
			},
		},
		Conditions: lb.ListenerRuleConditionArray{
			&lb.ListenerRuleConditionArgs{
				HostHeader: &lb.ListenerRuleConditionHostHeaderArgs{
					Values: pulumi.ToStringArray(e.Backend.Hostnames),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &Result{
		AlbArn:     albRes.Arn,
		AlbDnsName: albRes.DnsName,
		AlbZoneId:  albRes.ZoneId,

		HttpsListenerArn: httpsListener.Arn,

		FrontendTargetGroupArn: feTg.Arn,
		BackendTargetGroupArn:  beTg.Arn,

		FrontendRuleArn: feRule.Arn,
		BackendRuleArn:  beRule.Arn,
	}, nil
}
