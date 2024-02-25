// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package claims

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
)

func LegacyAudience(ctx context.Context, opt *FactoryOption) utils.StringSet {
	// in the java implementation, Spring uses "aud" for resource IDs which has been deprecated
	client, ok := ctx.Value(oauth2.CtxKeyAuthenticatedClient).(oauth2.OAuth2Client)
	if !ok || client.ResourceIDs() == nil || len(client.ResourceIDs()) == 0 {
		return utils.NewStringSet(oauth2.LegacyResourceId)
	}

	return client.ResourceIDs()
}
