# OAuth2 Auth Server

## List of Grants
| Grant Name         | Suitable for Public Client |
|--------------------|----------------------------|
| Client Credential  | No                         |
| Password           | No                         |
| Authorization Code | Yes                        |
| Refresh Token      | Yes                        |
| Switch Tenant      | Yes                        |
| Switch User        | Yes                        |

Public clients are clients that are considered not able to keep a secret (such as javascript code executing in the browser.) The grants
listed that are suitable for public clients relies on user authentication in addition to client authentication. You should consider the client authentication
for public clients to be untrusted because the client cannot store secrets reliably. 

## Client Credential
This grant allows a client to authenticate itself using its clientId and client secret. The authorization server would return
an access token upon authentication. The request can include an optional tenant id parameter to select the tenant for the
resulting security context. If no tenant id is provided, the resulting tenancy for the security context would be based on the
 calculation for default tenant. If the client only have one assigned tenant, it will be used as the default. If the client have multiple
assigned tenant, the security context will not have any tenancy. In that case, the caller must specify the tenant id to select
a tenant if tenancy is desired.

### Fields
| Field         | Value                             | Note                      |
|---------------|-----------------------------------|---------------------------|
| Method        | POST                              |                           |
| Target        | /v2/token                         |                           |
| grant_type    | client_credentials                | url values                |
| tenant_id     | tenant id                         | optional url values       |
| Content-Type  | application/x-www-form-urlencoded | request header            |
| Accept        | application/json                  | request header            |
| Authorization | Use the basic auth                | clientID:Secret in base64 |

### Curl Example
```bash
curl --location --request POST 'http://localhost:8900/auth/v2/token?grant_type=client_credentials' \
--header 'Authorization: Basic {base64_encode(clientId:clientSecret}'
```

## Password
This grant allows client to authenticate both the client and the user by issuing both the client id client secret and username and password.
The optional tenant id parameter will select the current tenant for the resulting security context. The tenants this authentication context
has access to is based on the intersection of the user's assigned tenants and the client's assigned tenants. This grant requires the authorization
server to be able to authenticate user using its password. It's not applicable if the user is authenticated via a SSO protocol (such as SAML).

### Fields
| Field         | Value                             | Note                      |
|---------------|-----------------------------------|---------------------------|
| Method        | POST                              |                           |
| Target        | /v2/token                         |                           |
| grant_type    | password                          | url values                |
| username      | username                          | url values                |
| password      | password                          | url values                |
| tenant_id     | tenant id                         | optional url values       |
| Content-Type  | application/x-www-form-urlencoded | request header            |
| Accept        | application/json                  | request header            |
| Authorization | Use the basic auth                | clientID:Secret in base64 |


### Curl Example

```bash
curl --location --request POST 'http://localhost:8900/auth/v2/token' \
--header 'Authorization: Basic {base64_encode(clientId:clientSecret}' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--header 'Accept: application/json' \
--data-urlencode 'grant_type=password' \
--data-urlencode 'username={username}' \
--data-urlencode 'password={password}' \
--data-urlencode 'tenant_id={tenant-id}'
```

## Authorization Code
This grant allows client to authenticate both the client and user. Unlike the password grant, it doesn't need the user to provide
their credentials to the client. The authorization request returns an auth code. The client needs to call the token API with the auth code
to get the access token. In the token request, the tenant id parameter is an optional parameter to select the tenant for the resulting security context. 
The tenants this authentication context has access to is based on the intersection of the user's assigned tenants and the client's assigned tenants.

### Authorize Request Fields
See OAuth2 Spec for definition of corresponding fields

| Field         | Value                             | Note                      |
|---------------|-----------------------------------|---------------------------|
| Method        | GET                               |                           |
| Target        | /v2/authorize                     |                           |
| response_type | code                              | url values                |
| client_id     | client id                         | url values                |
| redirect_uri  | redirect uri                      | url values                |
| state         | state                             | url values                |

### Example Request Issued from Browser

```
GET /auth/v2/authorize?response_type=code&client_id={client_id}}&redirect_uri={redirect_uri}&state={state} HTTP/1.1
Host: localhost:8900
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate
Connection: keep-alive
Cookie: SESSION=785e280f-4b1e-490d-b447-581ee357ddeb
Upgrade-Insecure-Requests: 1
Pragma: no-cache
Cache-Control: no-cache
```

### Token Request Fields
The response for this grant can also include a refresh token. See oauth2 spec on how to use the refresh token.

| Field         | Value                             | Note                      |
|---------------|-----------------------------------|---------------------------|
| Method        | POST                              |                           |
| Target        | /v2/token                         |                           |
| grant_type    | authorization_code                | url values                |
| code          | username                          | url values                |
| client_id     | client_id                         | url values                |
| redirect_uri  | redirect_uri                      | url values                |
| tenant_id     | tenant id                         | optional url values       |
| Content-Type  | application/x-www-form-urlencoded | request header            |
| Accept        | application/json                  | request header            |
| Authorization | Use the basic auth                | clientID:Secret in base64 |

### Curl Example

```bash
curl --location --request POST 'http://localhost:8900/auth/v2/token' \
--header 'Authorization: Basic {base64_encode(clientId:clientSecret}' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--header 'Accept: application/json' \
--data-urlencode 'grant_type=authorization_code' \
--data-urlencode 'code={code}' \
--data-urlencode 'client_id={client_id}' \
--data-urlencode 'redirect_uri={redirect_uri}'
```

## Switch User
Use an access token to switch to a different user, resulting in a new access token. The current user must be granted the permission to switch user.

### Fields

| Field           | Value                                      | Note                      |
|-----------------|--------------------------------------------|---------------------------|
| Method          | POST                                       |                           |
| Target          | /v2/token                                  |                           |
| grant_type      | urn:cisco:nfv:oauth:grant-type:switch-user | url values                |
| access_token    | access token value                         | url values                |
| switch_username | target user name                           | url values                |
| switch_user_id  | target user id                             | url values                |
| tenant_id       | tenant id                                  | url values                |
| Content-Type    | application/x-www-form-urlencoded          | request header            |
| Accept          | application/json                           | request header            |
| Authorization   | Use the basic auth                         | clientID:Secret in base64 |

### Curl Example

```bash
curl --location --request POST 'http://localhost:8900/auth/v2/token' \
--header 'Authorization: Basic {base64_encode(clientId:clientSecret}' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--header 'Accept: application/json' \
--data-urlencode 'grant_type=urn:cisco:nfv:oauth:grant-type:switch-user' \
--data-urlencode 'access_token={access_token}' \
--data-urlencode `switch_user_id={switch_user_id}` \
--data-urlencode 'tenant_id={tenant-id}'
```

## Switch Tenant
Use an access token to switch to a different tenant, resulting in a new access token. The current user must be granted the permission to switch tenant.

### Fields

| Field         | Value                                        | Note                       |
|---------------|----------------------------------------------|----------------------------|
| Method        | POST                                         |                            |
| Target        | /v2/token                                    |                            |
| grant_type    | urn:cisco:nfv:oauth:grant-type:switch-tenant | url values                 |
| access_token  | access token value                           | url values                 |
| tenant_id     | tenant id                                    | url values                 |
| Content-Type  | application/x-www-form-urlencoded            | request header             |
| Accept        | application/json                             | request header             |
| Authorization | Use the basic auth                           | clientID:Secret in base64  |

Note that tenant external ID is deprecated. Please use tenantID for the field/value

### Curl Example

```bash
curl --location --request POST 'http://localhost:8900/auth/v2/token' \
--header 'Authorization: Basic {base64_encode(clientId:clientSecret}' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--header 'Accept: application/json' \
--data-urlencode 'grant_type=urn:cisco:nfv:oauth:grant-type:switch-tenant' \
--data-urlencode 'access_token={access_token}' \
--data-urlencode 'tenant_id={tenant-id}'
```

## Client Scopes

| Scope            | Usage                                                                   |
|------------------|-------------------------------------------------------------------------|
| read             | not used                                                                |
| write            | not used                                                                |
| openid           | Client needs this scope to engage OIDC in addition to OAuth2            |
| profile          | OIDC scope to get user profile related claims in user info and id token |
| email            | OIDC scope to get user email related claims in user info and id token   |
| address          | OIDC scope to get user address related claims in user info and id token |
| phone            | OIDC scope to get user phone related claims in user info and id token   |
| token_details    | allows client to get token details from check_token API                 |
| tenant_hierarchy | allows client to use the tenant_hierarchy API                           |
| cross_tenant     | allows client to access all tenants                                     |

## Client Registration Consideration

Client registration should be implemented in by application. It is not implemented in the framework. It is the service
implementation's responsibility to restrict grant types and scopes to clients as appropriate.

For grant types, client registration should require the client to have a client secret before allowing giving it a grant type that's not 
suitable for public clients. In addition, the switch tenant grant may not be suitable for a client that is supposed to work under only one tenant.

| Grant Name         | Suitable for Public Client | Suitable for Self Registered Client               |
|--------------------|----------------------------|---------------------------------------------------|
| Client Credential  | No                         | Yes                                               |
| Password           | No                         | Yes                                               |
| Authorization Code | Yes                        | Yes                                               |
| Refresh Token      | Yes                        | Yes                                               |
| Switch Tenant      | Yes                        | Depends on if client is supposed to be per tenant |
| Switch User        | Yes                        | Yes                                               |

For scopes, client registration should consider whether a scope is suitable to be given to a customer created client. The tenant_hierarchy 
and cross_tenant scope should not be given to customer registered client that is supposed to work only within the context of a single tenant.

The token_details scope should not be given to a customer registered client because it's related to the introspection API (the check_token API), 
and this API is meant to be used by resource owners. In most cases self registered clients are not resource owners.

Client registration should also consider whether a scope should be given to a public client. The token_details and tenant_hierarchy scope should 
not be given to any public client because they can't be trusted with keeping client secret.

| Scope            | Suitable for Public Client | Suitable for Self Registered Client                        |                                             
|------------------|----------------------------|------------------------------------------------------------|
| read             | Yes                        | Yes                                                        |
| write            | Yes                        | Yes                                                        |
| openid           | Yes                        | Yes                                                        |
| profile          | Yes                        | Yes                                                        |
| email            | Yes                        | Yes                                                        |
| address          | Yes                        | Yes                                                        |
| phone            | Yes                        | Yes                                                        |
| token_details    | No                         | No                                                         |
| tenant_hierarchy | No                         | No (assuming self registered client is isolated to tenant) |                                                       |
| cross_tenant     | Yes                        | No (assuming self registered client is isolated to tenant) |

## Check Token
This API allows a client to check a given token's validity. In addition, a client with the token_details scope can get the security context
details represented by this token by specifying ```no_details=false``` 

### Fields

| Field            | Value                             | Note                      |
|------------------|-----------------------------------|---------------------------|
| Method           | POST                              |                           |
| Target           | /v2/check_token                   |                           |
| token            | token value                       | url values                |
| tenant_type_hint | access_token or refresh_token     | url values                |
| Content-Type     | application/x-www-form-urlencoded | request header            |
| Accept           | application/json                  | request header            |
| Authorization    | Use the basic auth                | clientID:Secret in base64 |

### Curl Example

```bash
curl --location --request POST 'http://localhost:8900/auth/v2/check_token' \
--header 'Authorization: Basic {base64_encode(clientId:clientSecret}=' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'token={access_token}' \
--data-urlencode 'token_type_hint=access_token' \
--data-urlencode 'no_details=true'
```