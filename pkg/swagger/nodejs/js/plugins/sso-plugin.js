/**
 * Copyright 2023 Cisco Systems, Inc. and its affiliates
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as ssoActions from "./sso-actions"
import SsoStandaloneLayout from "../components/layout";
import SsoTopBar from "../components/topbar";

export const PluginErrorSource = "SSO";

export const SsoStatus = {
    Unauthorized: 0,
    Authorizing: 1,
    RequestingToken: 2,
    Authorized: 3,
};

export const SsoStateType = {
    Configure: "sso_configure",
    Init: "sso_init",
    Status: "sso_status",
    Token: "sso_token",
}

export const NfvOAuth2SsoPlugin = () => {
    return {
        afterLoad(system) {
            this.rootInjects.initSso = (configs) => {
                system.ssoActions.configureSso(configs);
                system.ssoActions.startOrResumeAuthorize(system);
            }
        },
        statePlugins: {
            sso: {
                actions: ssoActions,
                reducers: {
                    [SsoStateType.Status]: (state, { payload }) => {
                        return state.merge(payload)
                    },
                    [SsoStateType.Token]: (state, { payload }) => {
                        const {token} = payload;
                        state = state.merge(payload);
                        if (!token) {
                            return state.delete('token')
                        }
                        return state;
                    },
                    [SsoStateType.Configure]: (state, { payload }) => {
                        return state.set("configs", payload)
                    },
                    [SsoStateType.Init]: (state, { payload }) => {
                        return state.merge(payload)
                    },
                },
                selectors: {
                    state: (state) => state,
                    ssoConfigs: (state) => state.get("configs"),
                    status: (state) => state.get("status"),
                    token: (state) => state.get("token"),
                    isAuthorizing: (state) => state.has("status") && state.get("status") > SsoStatus.Unauthorized && state.get("status") < SsoStatus.Authorized,
                    isAuthorized: (state) => state.has("status") && state.has("token") && state.get("status") === SsoStatus.Authorized,
                    hasAccessToken: (state) => state.hasIn(["token", "access_token"]),
                    hasRefreshToken: (state) => state.hasIn(["token", "refresh_token"]),
                    isTokenExpired: (state) => {
                        if (!state.hasIn(["token", "expireTime"])) {
                            return true;
                        }
                        const expireTime = state.getIn(["token", "expireTime"])
                        return !expireTime || Date.now() >= expireTime.getTime();
                    },
                    getAccessToken: (state) => state.getIn(["token", 'access_token']),
                    getRefreshToken: (state) => state.getIn(["token", 'refresh_token']),
                    getFromTokenResponse: (state, path) => {
                        let pathArray = path.split(",")
                        return state.getIn(["token", ...pathArray])
                    },
                    originalRequestInterceptor: (state) => state.get("originalRequestInterceptor"),
                    originalResponseInterceptor: (state) => state.get("originalResponseInterceptor"),
                },
            },
            configs: {
                wrapActions: {
                    loaded: (original, system) => () => {
                        ssoActions.initSsoInterceptors(system);
                        return original();
                    },

                }
            },
        },
        components: {
            SsoStandaloneLayout: SsoStandaloneLayout,
            SsoTopBar: SsoTopBar
        },
    }
};