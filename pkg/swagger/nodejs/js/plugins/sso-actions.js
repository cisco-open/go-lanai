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

import parseUrl from "url-parse"
import {buildFormData} from "../utils";
import {authorize, processAuthCodeResult, loadRedirectProgress, resetRedirectContext} from "./sso-authorize"
import {SsoStatus, SsoStateType, PluginErrorSource} from "./sso-plugin";

/**
 * Helper functions
 */
const ssoRequestInterceptor = ({ssoSelectors, ssoActions}) => (request) => {
    const original = ssoSelectors.originalRequestInterceptor();
    if (original) {
        request = original(request);
    }

    if (ssoSelectors.isAuthorized() && ssoSelectors.hasAccessToken()) {
        if (!ssoSelectors.isTokenExpired()) {
            if (!request.headers['Authorization'] ) {
                request.headers['Authorization'] = "Bearer " + ssoSelectors.getAccessToken();
            } else {
                console.warn('SSO access token is not applied, because there is "Authorization" header in request')
            }
        } else {
            ssoActions.accessTokenExpired();
        }
    }
    return request
};

const ssoResponseInterceptor = ({ssoSelectors}) => (response) => {
    const original = ssoSelectors.originalResponseInterceptor();
    if (original) {
        response = original(response);
    }
    return response;
};


/**
 * Actions
 */
export function configureSso(payload) {
    console.info("Configuring SSO");
    return {
        type: SsoStateType.Configure,
        payload: payload
    }
}

export function initSsoInterceptors(system) {
    console.info("Configuring Interceptors");
    const { getConfigs } = system;
    const configs = getConfigs();
    const payload = {
        originalRequestInterceptor: configs.requestInterceptor,
        originalResponseInterceptor: configs.responseInterceptor,
    };
    configs.requestInterceptor = ssoRequestInterceptor(system);
    configs.responseInterceptor = ssoResponseInterceptor(system);

    return {
        type: SsoStateType.Init,
        payload: payload,
    }
}

export function startOrResumeAuthorize({ssoActions, errActions, ssoSelectors, parameterName, parameterValue}) {
    const ssoConfigs = ssoSelectors.ssoConfigs();
    const authorized = ssoSelectors.isAuthorized();
    const progress = loadRedirectProgress();

    if (ssoConfigs && !authorized && progress) {
        // Already tried to authorize, continue
        console.log("SSO Authorization Resuming...");
        if (processAuthCodeResult({result: progress, ssoActions, errActions, ssoConfigs}) ) {
            return {
                type: SsoStateType.Status,
                payload: { status: SsoStatus.Authorizing }
            }
        }
    } else {
        return ssoAuthorize({ssoActions, errActions, ssoSelectors, parameterName, parameterValue});
    }

    return {
        type: SsoStateType.Status,
        payload: { }
    }
}

export function ssoAuthorize({ssoActions, errActions, ssoSelectors, parameterName, parameterValue}) {
    const ssoConfigs = ssoSelectors.ssoConfigs();
    const authorized = ssoSelectors.isAuthorized();

    if (ssoConfigs) {
        console.info("SSO Authorizing with parameter name " + parameterName + " " + parameterValue);
        if (authorize({ssoActions, errActions, ssoConfigs, parameterName, parameterValue})) {
            return {
                type: SsoStateType.Status,
                payload: { status: SsoStatus.Authorizing }
            }
        }
    } else if (!ssoConfigs) {
        log.warn("SSO Configs is not available")
    }

    return {
        type: SsoStateType.Status,
        payload: { }
    }
}

// Note this function's result is not handled by reducers
export const ssoAuthorized = ( { code, redirectUrl } ) => ( system ) => {
    let { ssoActions, ssoSelectors } = system;
    let { tokenUrl, clientId, clientSecret } = ssoSelectors.ssoConfigs();
    const headers = {
        Authorization: "Basic " + btoa(clientId + ":" + clientSecret)
    }
    const form = {
        grant_type: "authorization_code",
        code: code,
        client_id: clientId,
        redirect_uri: redirectUrl
    }
    const data = {body: buildFormData(form), url: tokenUrl, headers};

    ssoActions.ssoRequestToken({data}, system)
};

// Note this function's result is not handled by reducers
export const accessTokenExpired = () => ( system ) => {
    let { ssoActions, ssoSelectors, errActions } = system;
    if (!ssoSelectors.isAuthorized()) {
        return;
    } else if (ssoSelectors.isAuthorized() && !ssoSelectors.hasRefreshToken()) {
        errActions.newAuthErr({
            authId: "SSO",
            level: "error",
            source: PluginErrorSource,
            message: "Refresh token is not allowed. Please refresh page."
        });
        return
    }

    let { tokenUrl, clientId, clientSecret } = ssoSelectors.ssoConfigs();
    const refreshToken = ssoSelectors.getRefreshToken();
    const headers = {
        Authorization: "Basic " + btoa(clientId + ":" + clientSecret)
    }
    const form = {
        grant_type: "refresh_token",
        refresh_token: refreshToken,
        client_id: clientId,
    }
    const data = {body: buildFormData(form), url: tokenUrl, headers};
    const errorHandler = (e) => {
        if (e && e.response && e.response.status === 401) {
            console.info("Failed to refresh token. Current session is probably timed out. Re-authorizing...")
            ssoActions.ssoAuthorize(system)
        }
    }

    console.info("Refreshing token");
    ssoActions.ssoRemoveToken({message: "Refresh token"}, system);
    ssoActions.ssoRequestToken({data, errorHandler}, system);
};

export function ssoRequestToken( { data, successHandler, errorHandler }, system ){
    let { fn, getConfigs, ssoActions, errActions, oas3Selectors, specSelectors } = system;
    let { body, query={}, headers={}, url } = data;
    const name = "SSO";
    let parsedUrl

    if (specSelectors.isOAS3()) {
        parsedUrl = parseUrl(url, oas3Selectors.selectedServer(), true)
    } else {
        parsedUrl = parseUrl(url, specSelectors.url(), true)
    }

    if(typeof additionalQueryStringParams === "object") {
        parsedUrl.query = Object.assign({}, parsedUrl.query, additionalQueryStringParams)
    }

    const fetchUrl = parsedUrl.toString()

    const _headers = Object.assign({
        "Accept":"application/json, text/plain, */*",
        "Content-Type": "application/x-www-form-urlencoded",
        "X-Requested-With": "XMLHttpRequest"
    }, headers)

    fn.fetch({
        url: fetchUrl,
        method: "post",
        headers: _headers,
        query: query,
        body: body,
        requestInterceptor: getConfigs().requestInterceptor,
        responseInterceptor: getConfigs().responseInterceptor
    })
        .then(function (response) {
            const token = JSON.parse(response.data)
            const error = token && ( token.error || "" )
            const parseError = token && ( token.parseError || "" )

            if ( !response.ok ) {
                errActions.newAuthErr( {
                    authId: name,
                    level: "error",
                    source: PluginErrorSource,
                    message: response.statusText
                } )
                return
            }

            if ( error || parseError ) {
                errActions.newAuthErr({
                    authId: name,
                    level: "error",
                    source: PluginErrorSource,
                    message: JSON.stringify(token)
                })
                return
            }

            ssoActions.ssoToken({ token }, system);
            successHandler && successHandler(response);
        })
        .catch(e => {
            const err = new Error(e)
            let message = err.message + " [";
            // swagger-js wraps the response (if available) into the e.response property;
            // investigate to check whether there are more details on why the authorization
            // request failed (according to RFC 6479).
            // See also https://github.com/swagger-api/swagger-ui/issues/4048
            if (e.response && e.response.data) {
                const errData = e.response.data
                try {
                    const jsonResponse = typeof errData === "string" ? JSON.parse(errData) : errData
                    if (jsonResponse.error) {
                        message += `${jsonResponse.error}`;
                    }
                    if (jsonResponse.error_description) {
                        message += ` - ${jsonResponse.error_description}`;
                    }
                    message += ']';
                } catch (jsonError) {
                    // Ignore
                }
            }
            errActions.newAuthErr( {
                authId: name,
                level: "error",
                source: PluginErrorSource,
                message: message
            } )
            ssoActions.ssoRemoveToken({ message }, system);
            errorHandler && errorHandler(e);
        });

    return {
        type: SsoStateType.Status,
        payload: { status: SsoStatus.RequestingToken }
    }
}

export function ssoToken( { token }, {ssoActions, errActions}) {
    resetRedirectContext();
    // console.debug(token)
    //calculate expire time
    let time;
    const {expires_in} = token;
    if (expires_in) {
        time = new Date(Date.now() + expires_in * 1000);
    } else {
        time = new Date('2099-12-31T23:59:59')
    }

    // record and schedule a refresh
    token.expireTime = time;
    token.refreshJob = setTimeout(() => {
        ssoActions.accessTokenExpired();
    }, time.getTime() - Date.now());

    // clear any previous error
    errActions.clear({source: PluginErrorSource})

    console.info("SSO Token Received")
    return {
        type: SsoStateType.Token,
        payload: { token, status: SsoStatus.Authorized }
    }
}

export function ssoRemoveToken( { reason }, {ssoSelectors}) {
    resetRedirectContext()
    const token = ssoSelectors.token();
    token && token.refreshJob && clearTimeout(token.refreshJob);

    console.info("SSO Token Removed")
    return {
        type: SsoStateType.Token,
        payload: { token: null, status: SsoStatus.Unauthorized }
    }
}

