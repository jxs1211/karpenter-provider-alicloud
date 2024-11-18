package unregisteredtaint

/*
Copyright 2024 The CloudPilot AI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
import (
	"context"
	"time"

	"github.com/awslabs/operatorpkg/singleton"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	v1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

// Controller is used to delete the unregistered taint when the node is ready
type Controller struct {
	kubeClient client.Client
}

func NewController(kubeClient client.Client) *Controller {
	return &Controller{
		kubeClient: kubeClient,
	}
}

func (c *Controller) Reconcile(ctx context.Context) (reconcile.Result, error) {
	logger := log.FromContext(ctx)
	// get all nodes
	nodeList := &corev1.NodeList{}
	if err := c.kubeClient.List(ctx, nodeList, client.HasLabels{v1.NodeRegisteredLabelKey}); err != nil {
		return reconcile.Result{}, err
	}
	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		if !hasUnregisteredTaint(node) || !isNodeReady(node) {
			continue
		}

		nodeCopy := node.DeepCopy()
		// remove the unregistered taint
		nodeCopy.Spec.Taints = lo.Reject(nodeCopy.Spec.Taints, func(item corev1.Taint, index int) bool {
			return item.MatchTaint(&v1.UnregisteredNoExecuteTaint)
		})

		if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return c.kubeClient.Patch(ctx, nodeCopy, client.MergeFromWithOptions(node, client.MergeFromWithOptimisticLock{}))
		}); err != nil {
			logger.Error(err, "failed to remove unregistered taint", "node", node.Name)
			continue
		}
		logger.Info("removed unregistered taint from node", "node", node.Name)
	}

	// check again every minute
	return reconcile.Result{RequeueAfter: time.Minute}, nil
}

func (c *Controller) Register(_ context.Context, m manager.Manager) error {
	return controllerruntime.NewControllerManagedBy(m).
		Named("node.unregisteredtaint").
		WatchesRawSource(singleton.Source()).
		Complete(singleton.AsReconciler(c))
}

func hasUnregisteredTaint(node *corev1.Node) bool {
	_, has := lo.Find(node.Spec.Taints, func(item corev1.Taint) bool {
		return item.Key == v1.UnregisteredTaintKey
	})
	return has
}

func isNodeReady(node *corev1.Node) bool {
	_, ready := lo.Find(node.Status.Conditions, func(item corev1.NodeCondition) bool {
		return item.Type == corev1.NodeReady &&
			item.Status == corev1.ConditionTrue &&
			time.Since(item.LastTransitionTime.Time) > time.Second*15
	})

	return ready
}
