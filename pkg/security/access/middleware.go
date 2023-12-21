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

package access

import (
	"github.com/gin-gonic/gin"
)

//goland:noinspection GoNameStartsWithPackageName
type AccessControlMiddleware struct {
	decisionMakers []DecisionMakerFunc
}

func NewAccessControlMiddleware(decisionMakers...DecisionMakerFunc) *AccessControlMiddleware {
	return &AccessControlMiddleware{decisionMakers: decisionMakers}
}

func (ac *AccessControlMiddleware) ACHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error
		for _, decisionMaker := range ac.decisionMakers {
			var handled bool
			handled, err = decisionMaker(ctx, ctx.Request)
			if handled {
				break
			}
		}

		if err != nil {
			// access denied
			ac.handleError(ctx, err)
		} else {
			ctx.Next()
		}
	}
}

func (ac *AccessControlMiddleware) handleError(c *gin.Context, err error) {
	// We add the error and let the error handling middleware to render it
	_ = c.Error(err)
	c.Abort()
}
