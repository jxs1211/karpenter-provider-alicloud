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

package ack

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	ackclient "github.com/alibabacloud-go/cs-20151215/v5/client"
	"github.com/alibabacloud-go/tea/tea"
	"sigs.k8s.io/controller-runtime/pkg/log"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"

	"github.com/cloudpilot-ai/karpenter-provider-alibabacloud/pkg/apis/v1alpha1"
)

type Provider interface {
	GetNodeRegisterScript(context.Context, map[string]string, *v1alpha1.KubeletConfiguration) (string, error)
}

type DefaultProvider struct {
	clusterID string
	ackClient *ackclient.Client
}

func NewDefaultProvider(clusterID string, ackClient *ackclient.Client) *DefaultProvider {
	return &DefaultProvider{
		clusterID: clusterID,
		ackClient: ackClient,
	}
}

func (p *DefaultProvider) GetClusterCNI(ctx context.Context) (string, error) {
	response, err := p.ackClient.DescribeClusterDetail(tea.String(p.clusterID))
	if err != nil {
		return "", fmt.Errorf("failed to describe cluster: %w", err)
	}
	if response.Body == nil {
		return "", fmt.Errorf("empty cluster response")
	}
	if response.Body.MetaData == nil {
		return "", fmt.Errorf("empty cluster metadata")
	}
	// Parse metadata JSON string
	// clusterMetaData represents the metadata structure in cluster response
	type clusterMetaData struct {
		Capabilities struct {
			Network string `json:"Network"`
		} `json:"Capabilities"`
	}
	var metadata clusterMetaData
	if err := json.Unmarshal([]byte(*response.Body.MetaData), &metadata); err != nil {
		return "", fmt.Errorf("failed to unmarshal cluster metadata: %w", err)
	}
	return metadata.Capabilities.Network, nil
}

func (p *DefaultProvider) GetNodeRegisterScript(ctx context.Context,
	labels map[string]string,
	kubeletCfg *v1alpha1.KubeletConfiguration) (string, error) {
	reqPara := &ackclient.DescribeClusterAttachScriptsRequest{
		KeepInstanceName: tea.Bool(true),
	}

	// TODO: Build a cache to store this
	resp, err := p.ackClient.DescribeClusterAttachScripts(tea.String(p.clusterID), reqPara)
	if err != nil {
		log.FromContext(ctx).Error(err, "Failed to get node registration script")
		return "", err
	}
	s := tea.StringValue(resp.Body)
	if s == "" {
		err := errors.New("empty node registration script")
		log.FromContext(ctx).Error(err, "")
		return "", err
	}

	return p.resolveUserData(s, labels, kubeletCfg), nil
}

func (p *DefaultProvider) resolveUserData(respStr string,
	labels map[string]string,
	kubeletCfg *v1alpha1.KubeletConfiguration) string {
	cleanupStr := strings.ReplaceAll(respStr, "\r\n", "")

	// TODO: now, the following code is quite ugly, make it clean in the future
	// Add labels
	labelsFormated := fmt.Sprintf("ack.aliyun.com=%s", p.clusterID)
	for labelKey, labelValue := range labels {
		labelsFormated = fmt.Sprintf("%s,%s=%s", labelsFormated, labelKey, labelValue)
	}
	re := regexp.MustCompile(`--labels\s+\S+`)
	updatedCommand := re.ReplaceAllString(cleanupStr, "--labels "+labelsFormated)

	// Add kubelet config
	cfg := convertNodeClassKubeletConfigToACKNodeConfig(kubeletCfg)
	updatedCommand = fmt.Sprintf("%s --node-config %s", updatedCommand, cfg)

	// Add taints
	taint := karpv1.UnregisteredNoExecuteTaint
	updatedCommand = fmt.Sprintf("%s --taints %s", updatedCommand, taint.ToString())

	// Add bash script header
	finalScript := fmt.Sprintf("#!/bin/bash\n\n%s", updatedCommand)

	return base64.StdEncoding.EncodeToString([]byte(finalScript))
}

type NodeConfig struct {
	KubeletConfig *ACKKubeletConfig `json:"kubelet_config,omitempty"`
}

type ACKKubeletConfig struct {
	MaxPods *int32 `json:"maxPods,omitempty"`
}

func convertNodeClassKubeletConfigToACKNodeConfig(kubeletCfg *v1alpha1.KubeletConfiguration) string {
	cfg := &NodeConfig{
		KubeletConfig: &ACKKubeletConfig{
			MaxPods: kubeletCfg.MaxPods,
		},
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return base64.StdEncoding.EncodeToString([]byte("{}"))
	}
	return base64.StdEncoding.EncodeToString(data)
}
