/*
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

package instance

import (
	"errors"
	"fmt"
	"time"

	ecsclient "github.com/alibabacloud-go/ecs-20140526/v4/client"
	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/cloudpilot-ai/karpenter-provider-alibabacloud/pkg/utils"
)

const (
	// Ref: https://api.aliyun.com/api/Ecs/2014-05-26/DescribeInstances
	InstanceStatusPending  = "Pending"
	InstanceStatusRunning  = "Running"
	InstanceStatusStarting = "Starting"
	InstanceStatusStopping = "Stopping"
	InstanceStatusStopped  = "Stopped"
)

// Instance is an internal data representation of either an ecsclient.DescribeInstancesResponseBodyInstancesInstance
// It contains all the common data that is needed to inject into the Machine from either of these responses
type Instance struct {
	CreationTime     time.Time         `json:"creationTime"`
	Status           string            `json:"status"`
	ID               string            `json:"id"`
	ImageID          string            `json:"imageId"`
	Type             string            `json:"type"`
	Region           string            `json:"region"`
	Zone             string            `json:"zone"`
	CapacityType     string            `json:"capacityType"`
	SecurityGroupIDs []string          `json:"securityGroupIds"`
	VSwitchID        string            `json:"vSwitchId"`
	Tags             map[string]string `json:"tags"`
}

func NewInstance(out *ecsclient.DescribeInstancesResponseBodyInstancesInstance) *Instance {
	creationTime, err := utils.ParseISO8601(*out.CreationTime)
	if err != nil {
		log.Log.Error(err, "Failed to parse creation time")
	}

	return &Instance{
		CreationTime:     creationTime,
		Status:           *out.Status,
		ID:               *out.InstanceId,
		ImageID:          *out.ImageId,
		Type:             *out.InstanceType,
		Region:           *out.RegionId,
		Zone:             *out.ZoneId,
		CapacityType:     utils.GetCapacityTypes(*out.SpotStrategy),
		SecurityGroupIDs: toSecurityGroupIDs(out.SecurityGroupIds),
		VSwitchID:        toVSwitchID(out.VpcAttributes),
		Tags:             toTags(out.Tags),
	}
}

func NewInstanceFromProvisioningGroup(out *ecsclient.CreateAutoProvisioningGroupResponseBodyLaunchResultsLaunchResult,
	req *ecsclient.CreateAutoProvisioningGroupRequest, region string) *Instance {

	var securityGroupIDs []string
	if len(req.LaunchConfiguration.SecurityGroupIds) > 0 {
		securityGroupIDs = lo.Map(req.LaunchConfiguration.SecurityGroupIds, func(securitygroup *string, _ int) string {
			return *securitygroup
		})
	} else {
		securityGroupIDs = []string{*req.LaunchConfiguration.SecurityGroupId}
	}

	tags := make(map[string]string, len(req.Tag))
	for i := range req.Tag {
		tags[*req.Tag[i].Key] = *req.Tag[i].Value
	}

	return &Instance{
		CreationTime:     time.Now(), // estimate the launch time since we just launched
		Status:           InstanceStatusPending,
		ID:               *out.InstanceIds.InstanceId[0],
		ImageID:          *req.LaunchConfiguration.ImageId,
		Type:             *out.InstanceType,
		Region:           region,
		Zone:             *out.ZoneId,
		CapacityType:     utils.GetCapacityTypes(*out.SpotStrategy),
		SecurityGroupIDs: securityGroupIDs,
		Tags:             tags,
	}
}

func toSecurityGroupIDs(securityGroups *ecsclient.DescribeInstancesResponseBodyInstancesInstanceSecurityGroupIds) []string {
	if securityGroups == nil {
		return []string{}
	}

	return lo.Map(securityGroups.SecurityGroupId, func(securitygroup *string, _ int) string {
		return *securitygroup
	})
}

func toVSwitchID(vpcAttributes *ecsclient.DescribeInstancesResponseBodyInstancesInstanceVpcAttributes) string {
	if vpcAttributes == nil {
		return ""
	}

	return *vpcAttributes.VSwitchId
}

func toTags(tags *ecsclient.DescribeInstancesResponseBodyInstancesInstanceTags) map[string]string {
	if tags == nil {
		return map[string]string{}
	}

	return lo.SliceToMap(tags.Tag, func(tag *ecsclient.DescribeInstancesResponseBodyInstancesInstanceTagsTag) (string, string) {
		return *tag.TagKey, *tag.TagValue
	})
}

type InstanceStateOperationNotSupportedError struct {
	error
}

func NewInstanceStateOperationNotSupportedError(instanceID string) *InstanceStateOperationNotSupportedError {
	return &InstanceStateOperationNotSupportedError{error: fmt.Errorf("instance(%s) state not supported for this operation", instanceID)}
}

func IsInstanceStateOperationNotSupportedError(err error) bool {
	if err == nil {
		return false
	}

	var errAsInstanceStateOperationNotSupportedError *InstanceStateOperationNotSupportedError
	return errors.As(err, &errAsInstanceStateOperationNotSupportedError)
}
