# weather-probe
Simple demo backend that checks weather in London via OpenWeatherAPI, CA on "/" path deployed to EKS in GitOps way.

# Overview

The application is called weather-backend and is written in Golang, supports several paths:
/ - returns HTML with current weather in London, CA
/ping - returns PONG
/health - returns JSON with status

Deployment instructions for EKS cluster with NGINX ingress, external secrets and argocd could be found [here](/https://github.com/s0rl0v/weather-infra).

THere is GitOps repository for [ArgoCD](/https://github.com/s0rl0v/weather-k8s-deployments) with manifests to deploy and continuously sync applications.

The app is containerized and pushed to ECR registry via GitHub Actions pipeline, then ArgoCD Image Updater (that continiously monitors application ECR repo) modifies the image to the new one for corresponding environment, thus triggering rollout.

# Deployment

## Prerequisites

[OpenWeather API key](/https://openweathermap.org/appid)
[AWS programmatic access](/https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys)

## Steps

1. Navigate to infrastructure repo and deploy according to [the provided README](/https://github.com/s0rl0v/weather-infra/blob/main/README.md):

**NOTE**: Write down the output - this will come handy in next steps.

2. Navigate to [ArgoCD deployments](/https://github.com/s0rl0v/weather-k8s-deployments/blob/main/README.md) repo and review IAM roles, repositories URLs (if you want to use different that defined here), helm values (especially ingress for weather-backend), replace IAM roles in annotations (search by arn) with values aquired in step 1.

3. Add your API key to SSM:

```
aws ssm put-parameter --name "/sorlov/weather/OWM_API_KEY" --type "String" --value "******"
```

3. Navigate to ArgoCD deployment repo and apply manifests [README](/https://github.com/s0rl0v/weather-infra/blob/main/README.md).

4. Point application DNS name to nginx controller:

```
k get svc -n ingress-nginx ingress-nginx-controller
```

5. Test deployment by commiting something to the repo.


## Cleanup

1. Remove all the ArgoCD applicatons first (execute from ArgoCD deployments repo):

```
cd apps-registry && kustomize build | k delete -f-
```

2. Execute terraform destroy.

