import { sanitizeUrl } from "@braintree/sanitize-url"
import {btoa} from "../utils"
import { PluginErrorSource } from "./sso-plugin";

export const InProgressRedirectKey = "CurrentRedirectSSO";
export const SsoStateKeyPrefix = "sso-";

export function authorize ( { ssoActions, errActions, ssoConfigs={} , parameterName, parameterValue} ) {

    if (!validateSsoConfigs({errActions, ssoConfigs})) {
        return;
    }

    let { authorizeUrl, scopes, clientId, ssoRedirectUrl, usePopup=false } = ssoConfigs
    const query = []

    query.push("response_type=code");

    if (typeof clientId === "string") {
        query.push("client_id=" + encodeURIComponent(clientId))
    }

    const redirectUrl = ssoRedirectUrl;
    query.push("redirect_uri=" + encodeURIComponent(redirectUrl))

    if (Array.isArray(scopes) && 0 < scopes.length) {
        const scopeSeparator = ssoConfigs.scopeSeparator || " "
        query.push("scope=" + encodeURIComponent(scopes.join(scopeSeparator)))
    }

    const state = btoa(new Date())

    query.push("state=" + encodeURIComponent(state))

    if (typeof ssoConfigs.realm !== "undefined") {
        query.push("realm=" + encodeURIComponent(ssoConfigs.realm))
    }

    const { additionalQueryStringParams } = ssoConfigs

    for (const key in additionalQueryStringParams) {
        if (typeof additionalQueryStringParams[key] !== "undefined") {
            query.push([key, additionalQueryStringParams[key]].map(encodeURIComponent).join("="))
        }
    }

    if (parameterName && parameterValue) {
        query.push(parameterName+"="+parameterValue)
    }

    const authorizationUrl = authorizeUrl
    const sanitizedAuthorizationUrl = sanitizeUrl(authorizationUrl)
    const url = [sanitizedAuthorizationUrl, query.join("&")].join(authorizationUrl.indexOf("?") === -1 ? "?" : "&")

    // pass action authorizeOauth2 and authentication data through window
    // to authorize with oauth2

    let {callback, errCallback} = getCallbacks({ssoActions, errActions});

    let redirectCtx = {
        sso: {},
        state: state,
        redirectUrl: redirectUrl,
        callback: callback,
        errCallback: errCallback,
        originalUrl: window.location.href
    };
    updateRedirectContext({state, redirectCtx});

    const newWindow = usePopup ? window.open(url, 'nfv-swagger-login') : window.open(url, '_self');
    if (!newWindow) {
        updateRedirectContext({});
        errCallback({
            authId: "SSO",
            source: PluginErrorSource,
            level: "error",
            message: "Login popup was blocked."
        })
    } else {
        errActions.clear({source: PluginErrorSource})
    }

    return newWindow;
}

export function processAuthCodeResult({ result, ssoActions, errActions, ssoConfigs={} }) {
    if (!validateSsoConfigs({errActions, ssoConfigs})) {
        return;
    }

    let {callback, errCallback} = getCallbacks({ssoActions, errActions});
    resetRedirectContext();

    if (!result || !result.code || !result.isValid) {
        const level = result.error && result.error.level ? result.error.level : "error";
        const message = result.error && result.error.message ? result.error.message : "[Authorization failed]: Error when processing redirect callback"
        setTimeout(() => errCallback({
            authId: "SSO",
            source: PluginErrorSource,
            level: level,
            message: message
        }));
        return false;
    }

    setTimeout(() => callback({code: result.code, redirectUrl: result.redirectUrl}));
    return true;
}

export function loadRedirectProgress() {
    let redirectCtx = getStorage().getItem(InProgressRedirectKey);
    if (redirectCtx) {
        redirectCtx = JSON.parse(redirectCtx);
    }

    if (!redirectCtx || !redirectCtx.state) {
        return null;
    }

    // load auth code
    let key = `${SsoStateKeyPrefix}${redirectCtx.state}`;
    let progress = getStorage().getItem(key);
    return progress ? JSON.parse(progress) : null;
}

export function resetRedirectContext() {
    delete window.swaggerUIRedirectContext;

    let redirectCtx = getStorage().getItem(InProgressRedirectKey);
    getStorage().removeItem(InProgressRedirectKey);

    if (redirectCtx) {
        redirectCtx = JSON.parse(redirectCtx);
    }

    if (redirectCtx && redirectCtx.state) {
        let key = `${SsoStateKeyPrefix}${redirectCtx.state}`;
        getStorage().removeItem(key);
    }
}

function validateSsoConfigs({errActions, ssoConfigs}) {
    let { authorizeUrl, tokenUrl, clientId, clientSecret, ssoRedirectUrl } = ssoConfigs;

    const missingProps = [];
    !authorizeUrl && missingProps.push("authorizeUrl")
    !tokenUrl && missingProps.push("tokenUrl")
    !clientId && missingProps.push("clientId")
    !ssoRedirectUrl && missingProps.push("ssoRedirectUrl")
    if (missingProps.length !== 0) {
        errActions.newAuthErr( {
            authId: "SSO",
            source: PluginErrorSource,
            level: "error",
            message: 'SSO plugin is not configured properly. Missing required properties' + JSON.stringify(missingProps)
        })
        return false
    }
    return true;
}

function saveRedirectContext({redirectCtx}) {
    window.swaggerUIRedirectContext = redirectCtx;
    if (redirectCtx) {
        // to load function with from stringified JSON, we need to use eval. SonarQube doesn't like it
        // In fact, we don't need to store it in session, storing it in current window is sufficient.
        // Login in same window doesn't requires callback function
        delete redirectCtx.callback;
        delete redirectCtx.errCallback;
        getStorage().setItem(InProgressRedirectKey, JSON.stringify(redirectCtx));
    }
}

function updateRedirectContext({redirectCtx}) {
    if (redirectCtx) {
        saveRedirectContext({redirectCtx})
    } else {
        resetRedirectContext();
    }
}

function getCallbacks({ssoActions, errActions}) {
    return {
        callback: ssoActions.ssoAuthorized,
        errCallback: errActions.newAuthErr,
    }
}

function getStorage() {
    return window.sessionStorage;
}