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

package imagefamily

import (
	"regexp"
)

var (
	// The format should look like aliyun_3_arm64_20G_alibase_20240819.vhd
	alibabaCloudLinux3ImageIDRegex = regexp.MustCompile("aliyun_3_.*G_alibase_.*vhd")
)

type AlibabaCloudLinux3 struct{}

func (a *AlibabaCloudLinux3) ResolveImages(images Images) Images {
	var ret Images
	for _, im := range images {
		if !alibabaCloudLinux3ImageIDRegex.Match([]byte(im.ImageID)) {
			continue
		}

		ret = append(ret, im)
	}

	return ret
}
