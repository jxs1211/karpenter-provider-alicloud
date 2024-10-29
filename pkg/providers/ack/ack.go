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
	"errors"
	"fmt"
	"regexp"
	"strings"

	ackclient "github.com/alibabacloud-go/cs-20151215/v5/client"
	"github.com/alibabacloud-go/tea/tea"
	"sigs.k8s.io/controller-runtime/pkg/log"
	karpv1 "sigs.k8s.io/karpenter/pkg/apis/v1"
)

type Provider interface {
	GetNodeRegisterScript(ctx context.Context, labels map[string]string) (string, error)
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

func (p *DefaultProvider) GetNodeRegisterScript(ctx context.Context, labels map[string]string) (string, error) {
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

	return p.resolveUserData(s, labels), nil
}

func (p *DefaultProvider) resolveUserData(respStr string, labels map[string]string) string {
	cleanupStr := strings.ReplaceAll(respStr, "\r\n", "")

	// Add labels
	labelsFormated := fmt.Sprintf("ack.aliyun.com=%s", p.clusterID)
	for labelKey, labelValue := range labels {
		labelsFormated = fmt.Sprintf("%s,%s=%s", labelsFormated, labelKey, labelValue)
	}
	re := regexp.MustCompile(`--labels\s+\S+`)
	updatedCommand := re.ReplaceAllString(cleanupStr, "--labels "+labelsFormated)

	// Add taints
	taint := karpv1.UnregisteredNoExecuteTaint
	updatedCommand = fmt.Sprintf("%s --taints %s", updatedCommand, taint.ToString())

	// Add bash script header
	finalScript := fmt.Sprintf("#!/bin/bash\n\n%s", updatedCommand)

	return base64.StdEncoding.EncodeToString([]byte(finalScript))
}
