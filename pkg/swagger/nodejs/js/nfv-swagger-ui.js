import './csrf'
import SwaggerUI from 'swagger-ui'
import {
  NfvOAuth2SsoPlugin
} from './plugins/sso-plugin'


window.onload = () => {

  const getBaseURL = () => {
    const urlMatches = /(.*)\/swagger.*/.exec(window.location.href);
    return urlMatches[1];
  };

  const getUI = (baseUrl, resources, configUI, oauthSecurity, ssoSecurity) => {

    let layout = "StandaloneLayout";
    let plugins = [SwaggerUI.plugins.DownloadUrl];

    if (ssoSecurity && ssoSecurity.enabled) {
      layout = "SsoStandaloneLayout";
      plugins = [ SwaggerUI.plugins.DownloadUrl, NfvOAuth2SsoPlugin ];
    }

    const ui = SwaggerUI({
      /*--------------------------------------------*\
       * Core
      \*--------------------------------------------*/
      dom_id: "#swagger-ui",
      urls: resources,
      // configUrl: null,
      // dom_node: null,
      // spec: {},
      // url: "",
      /*--------------------------------------------*\
       * Plugin system
      \*--------------------------------------------*/
      layout: layout,
      plugins: plugins,
      presets: [
        SwaggerUI.presets.apis,
      ],
      /*--------------------------------------------*\
       * Display
      \*--------------------------------------------*/
      deepLinking: configUI.deepLinking,
      displayOperationId: configUI.displayOperationId,
      defaultModelsExpandDepth: configUI.defaultModelsExpandDepth,
      defaultModelExpandDepth: configUI.defaultModelExpandDepth,
      defaultModelRendering: configUI.defaultModelRendering,
      displayRequestDuration: configUI.displayRequestDuration,
      docExpansion: configUI.docExpansion,
      filter: configUI.filter,
      maxDisplayedTags: configUI.maxDisplayedTags,
      operationsSorter: configUI.operationsSorter,
      showExtensions: configUI.showExtensions,
      tagSorter: configUI.tagSorter,
      /*--------------------------------------------*\
       * Network
      \*--------------------------------------------*/
      oauth2RedirectUrl: baseUrl + "/swagger/static/oauth2-redirect.html",
      requestInterceptor: (a => a),
      responseInterceptor: (a => a),
      showMutatedRequest: true,
      supportedSubmitMethods: configUI.supportedSubmitMethods,
      validatorUrl: configUI.validatorUrl,
      /*--------------------------------------------*\
       * Macros
      \*--------------------------------------------*/
      modelPropertyMacro: null,
      parameterMacro: null,
    });

    oauthSecurity && ui.initOAuth({
      /*--------------------------------------------*\
       * OAuth
      \*--------------------------------------------*/
      clientId: oauthSecurity.clientId,
      clientSecret: oauthSecurity.clientSecret,
      realm: oauthSecurity.realm,
      appName: oauthSecurity.appName,
      scopeSeparator: oauthSecurity.scopeSeparator,
      additionalQueryStringParams: oauthSecurity.additionalQueryStringParams,
      useBasicAuthenticationWithAccessCodeGrant: oauthSecurity.useBasicAuthenticationWithAccessCodeGrant,
    });

    ssoSecurity && ssoSecurity.enabled && ui.initSso({
      /*--------------------------------------------*\
       * OAuth
      \*--------------------------------------------*/
      authorizeUrl: ssoSecurity.authorizeUrl,
      tokenUrl: ssoSecurity.tokenUrl,
      ssoRedirectUrl: baseUrl + "/swagger-sso-redirect.html",
      clientId: ssoSecurity.clientId,
      clientSecret: ssoSecurity.clientSecret,
    });

    return ui;
  };

  const buildSystemAsync = async (baseUrl) => {
    try {
      const configUIResponse = await fetch(
          baseUrl + "/swagger-resources/configuration/ui",
          {
            credentials: 'same-origin',
            headers: {
              'Accept': 'application/json',
              'Content-Type': 'application/json'
            },
          });
      const configUI = await configUIResponse.json();

      const configOAuth2SecurityResponse = await fetch(
          baseUrl + "/swagger-resources/configuration/security",
          {
            credentials: 'same-origin',
            headers: {
              'Accept': 'application/json',
              'Content-Type': 'application/json'
            },
          });
      const oauthSecurity = await configOAuth2SecurityResponse.json();

      const configSsoSecurityResponse = await fetch(
          baseUrl + "/swagger-resources/configuration/security/sso",
          {
            credentials: 'same-origin',
            headers: {
              'Accept': 'application/json',
              'Content-Type': 'application/json'
            },
          });
      const ssoSecurity = await configSsoSecurityResponse.json();

      const resourcesResponse = await fetch(
          baseUrl + "/swagger-resources",
          {
            credentials: 'same-origin',
            headers: {
              'Accept': 'application/json',
              'Content-Type': 'application/json'
            },
          });
      const resources = await resourcesResponse.json();
      resources.forEach(resource => {
        if (resource.url.substring(0, 4) !== 'http') {
          resource.url = baseUrl + resource.url;
        }
      });
      window.ui = getUI(baseUrl, resources, configUI, oauthSecurity, ssoSecurity);
    } catch (e) {
      console.error(e)
      const retryURL = await prompt(
        "Unable to infer base url. This is common when using dynamic servlet registration or when" +
        " the API is behind an API Gateway. The base url is the root of where" +
        " all the swagger resources are served. For e.g. if the api is available at http://example.org/api/v2/api-docs" +
        " then the base url is http://example.org/api/. Please enter the location manually: ",
        window.location.href);
      console.log(retryURL)
      // return buildSystemAsync(retryURL);
    }
  };

  /* Entry Point */
  (async () => {
    await buildSystemAsync(getBaseURL());
    // await csrfSupport(getBaseURL());
  })();

};
