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

package client

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewClientConfig(ctx context.Context, region string, network string) (*openapi.Config, error) {
	// Load in the following order: 1. AK/SK, 2. RRSA, 3. config.json, 4. RAMRole
	// https://www.alibabacloud.com/help/zh/sdk/developer-reference/v2-manage-go-access-credentials#3ca299f04bw3c
	credential, err := credentials.NewCredential(nil)
	if err != nil {
		return nil, err
	}
	if cred, err := credential.GetCredential(); err == nil && cred != nil {
		log.FromContext(ctx).Info(fmt.Sprintf("using credential type: %s, AccessKeyID: %s", tea.StringValue(cred.Type), tea.StringValue(cred.AccessKeyId)))
	} else {
		return nil, fmt.Errorf("failed get credential, error: %w", err)
	}

	return &openapi.Config{
		RegionId:   tea.String(region),
		Credential: credential,
		Network:    tea.String(network), // 1. public, 2. vpc, default is public
	}, nil
}
