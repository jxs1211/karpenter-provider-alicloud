# karpenter

A Helm chart for Karpenter, an open-source node provisioning project built for Kubernetes.

## Documentation

For the full Karpenter documentation, please check out [Karpenter-alibabacloud](https://docs.cloudpilot.ai/karpenter/alibabacloud/).

## Installing the Chart

```bash
export ALIBABACLOUD_AK=<your-access-key>
export ALIBABACLOUD_SK=<your-secret-key>
export CLUSTER_REGION=<your-region-id>
export CLUSTER_NAME=<your-cluster-name>
export CLUSTER_ID=<your-cluster-id>

helm upgrade karpenter ./charts/karpenter --install \
  --namespace karpenter-system --create-namespace \
  --set "alibabacloud.access_key_id"=${ALIBABACLOUD_AK} \
  --set "alibabacloud.access_key_secret"=${ALIBABACLOUD_SK} \
  --set "alibabacloud.region_id"=${ALIBABACLOUD_AK} \
  --set controller.settings.clusterName=${CLUSTER_NAME} \
  --set controller.settings.clusterID=${CLUSTER_ID} \
  --wait
```

## Uninstalling the Chart

```bash
helm uninstall karpenter --namespace karpenter-system
kubectl delete ns karpenter-system
```
