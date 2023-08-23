# Description

`ns-controller` is a validating admission webhook. Namespace create requests are sent to this webhook for validation. The purpose of this controller is two prevent a cluster denial of service because of uncontrolled workload submission.  
The controller calculates the total and used memory in the cluster. And based on those values and the user-defined overprovisioning coefficient it will allow or deny the admission request for the namespace creation.

To calculate those memory data the controller is deployed with a kube-state-metrics pod, which listens to the Kubernetes API server and generates metrics about the state of the objects.

# How does it work

`ns-controller` calculates:
- pod memory limits on all nodes
- all allocatable memory on all nodes

The formula `(pod memory limits sum) / (allocatable memory sum)` will give you an overprovisioning value.  
If it is smaller than the user-defined overprovisioning coefficient then the webhook will allow the namespace creation and not otherwise.

### Example:
The user-defined overprovisioning coefficient = `1.3`. At the moment of a new namespace creation the `sum of all pod memory limits` = 2325741568 bytes, and the `sum of all allocatable memory on nodes` = 1813823488 bytes. Then the overprovisioning value = `2325741568 / 1813823488 = 1.28223147587777` which is smaller than `1.3`, and the controller allows the admission request.

# How to use it

The controller uses the `cert-generator` repository for cert generation for the webhook server, see the docs at https://github.com/rv-0451/cert-generator

The helm chart values in `helm/ns-controller/values.yaml`:

| Name | Description |
| :---: | :---: |
| image.repository | registry where the image will be pushed |
| image.tag | tag of the image |
| init.repository | registry where the init container is stored |
| init.tag | tag of the init container |
| init.secretName | secret name where certificate data will be stored |
| webhookConfig.webhookNsName | the name of the webhook |
| webhookConfig.overprovisioning | overprovisioning coefficient |

# How to build and deploy

First, you need to build and push to your registry the `cert-generator` for webhook certificate configuration, see the docs at https://github.com/rv-0451/cert-generator

Update registry and tag values if necessary in:
- `Makefile`
- `helm/ns-controller/values.yaml`
- `helm/ns-controller/Chart.yaml`

To build and push the controller image:

```bash
make docker-build docker-push
```

To deploy the controller (you should be authorized to your cluster and have the helm binary on your host machine):

```bash
make helm-deploy
```

And to remove the controller from the cluster:

```bash
make helm-undeploy
```

For full actions see

```bash
make help
```
