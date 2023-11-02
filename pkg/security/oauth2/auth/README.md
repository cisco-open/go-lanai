# OAuth2 Auth Server

## Switch Tenant

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

## Check Token
