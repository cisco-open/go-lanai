# SAML SP

This  module enables a service to act as an SP (allows login with third party using SAML protocol). This feature has two
feature configurers.

**login feature configurer** does the following:

1. Add metadata endpoint (/saml/metadata)
2. Add ACS endpoint (/saml/SSO)
3. Add metadata refresh middleware that covers the above two endpoints
4. Make the metadata endpoint and acs endpoint public
5. Add an authentication entry point that will trigger the saml login process

**logout feature configurer** does the following:

1. Add single logout endpoint
2. Add metadata refresh middleware that covers the endpoint
3. Add logout handler
4. Add logout entry point (the entry point to send out the logout request to the IDP)

When SAML login feature is enabled, these middleware and endpoints are added to the web security configuration.


## Misc

Create saml private key and cert using the following command

```shell
openssl genrsa -out saml.key -aes256 1024
```

```shell
openssl req -key saml.key -new -x509 -days 36500 -out saml.crt
```