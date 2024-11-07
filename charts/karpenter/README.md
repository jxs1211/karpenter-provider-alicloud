# karpenter

A Helm chart for Karpenter, an open-source node provisioning project built for Kubernetes.

## Documentation

For the full Karpenter documentation, please check out [Karpenter-alibabacloud](https://docs.cloudpilot.ai/karpenter/alibabacloud/).

## Installing the Chart

```bash
helm upgrade karpenter ./charts/karpenter --install \
  --namespace karpenter --create-namespace \
  --set "alibabacloud.profiles[0].access_key_id"="<your-access-key>" \
  --set "alibabacloud.profiles[0].access_key_secret"="<your-secret-key>" \
  --set "alibabacloud.profiles[0].region_id"="<your-region-id>" \
  --set controller.settings.clusterName="<your-cluster-name>" \
  --set controller.settings.clusterID="<your-cluster-id>" \
  --wait
```

## Uninstalling the Chart

```bash
helm uninstall karpenter --namespace karpenter
kubectl delete ns karpenter
```
