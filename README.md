# fullstack-traba-copy-cat

```
This diagram shows the complete AWS infrastructure with several key components:

External Access Layer:

Route53 for DNS management
ACM certificate for SSL/TLS
Internet Gateway for public access


VPC Network (10.0.0.0/16):

Public Subnet (10.0.1.0/24):

Frontend and Backend ALBs
NAT Gateway for private subnet connectivity


Private Subnet (10.0.2.0/24):

ECS tasks running frontend and backend services




Security:

Layered security groups:

Frontend ALB SG (ports 80, 443)
Backend ALB SG (port 443)
Frontend Service SG (port 80)
Backend Service SG (port 3000)




Application Layer:

ECS Cluster running Fargate tasks
Frontend and Backend services in private subnet
Load balancers in public subnet


Monitoring:

CloudWatch Log Groups for both services
30-day log retention



Key security features:

Private subnet for application containers
Public subnet only for load balancers
Restricted security group access
HTTPS enforcement with automatic HTTP to HTTPS redirect
```

task deploy:pulumi PROJECT=fs0ciety ENV=prod

be aware the the terraform CI/CD are inthe .github/workflows-archived. if you want ot use them in CI then you have to move those file back into githubworkflows 
