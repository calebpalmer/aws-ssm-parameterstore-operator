# aws-ssm-operator

A kubernetes operator that creates secrets from AWS SSM Parameter Store parameters.

## TL;DR:
```bash
helm install aws-ssm-operator --set operator.awsAccessKeyId=123 --set operator.awsSecretAccessKey=*** --set operator.awsRegion=ca-central1
```

## Configuration

The following tables lists the configurable parameters of the alb-ssm-operator chart and their default values.

| Parameter                     | Description                                                                                                    | Default                                                                   |
| -------------------------     | -------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `operator.awsAccessKeyId`     | AWS Access Key Id of k8s cluster, required if ec2metadata is unavailable from controller pod                   |                                                                           |
| `oeprator.awsSecretAccessKey` | AWS Secret Access Key of k8s cluster, required if ec2metadata is unavailable from controller pod               |                                                                           |
| `operator.awsRegion`          | AWS region of k8s cluster, required if ec2metadata is unavailable from controller pod                          | `us-east-1 `                                                              |
| `image.repository`            | controller container image repository                                                                          | `calebpalmer/ssmparameter-operator`                                       |
| `image.pullPolicy`            | controller container image pull policy                                                                         | `IfNotPresent`                                                            |
| `nodeSelector`                | node labels for controller pod assignment                                                                      | `{}`                                                                      |
| `tolerations`                 | controller pod toleration for taints                                                                           | `{}`                                                                      |
| `podAnnotations`              | annotations to be added to controller pod                                                                      | `{}`                                                                      |
| `podLabels`                   | labels to be added to controller pod                                                                           | `{}`                                                                      |
| `priorityClassName`           | set to ensure your pods survive resource shortages                                                             | `""`                                                                      |
| `resources`                   | controller pod resource requests & limits                                                                      | `{}`                                                                      |

