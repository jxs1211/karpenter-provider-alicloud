# karpenter

A Helm chart for Karpenter, an open-source node provisioning project built for Kubernetes.

## Documentation

For the full Karpenter documentation, please check out [here](https://docs.cloudpilot.ai/karpenter/alibabacloud/).

## Installing the Chart

Run the following command the set the required environment variables before installing the chart:

```bash
export ALIBABACLOUD_AK=<your-access-key>
export ALIBABACLOUD_SK=<your-secret-key>
export CLUSTER_REGION=<your-region-id>
export CLUSTER_ID=<your-cluster-id>
```

Then run the following command to install the chart:

```bash
helm repo add karpenter-provider-alibabacloud https://cloudpilot-ai.github.io/karpenter-provider-alibabacloud

helm upgrade karpenter karpenter-provider-alibabacloud/karpenter --install \
  --namespace karpenter-system --create-namespace \
  --set "alibabacloud.access_key_id"=${ALIBABACLOUD_AK} \
  --set "alibabacloud.access_key_secret"=${ALIBABACLOUD_SK} \
  --set "alibabacloud.region_id"=${CLUSTER_REGION} \
  --set controller.settings.clusterID=${CLUSTER_ID} \
  --wait
```

## Uninstalling the Chart

```bash
helm uninstall karpenter --namespace karpenter-system
kubectl delete ns karpenter-system
```
