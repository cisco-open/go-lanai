# JWT

The JWT package provides support for working with JSON Web Tokens. This includes encoding and decoding tokens, as well as signing
and verifying the signature of tokens.

## JWK Store
In order to sign token (during encoding) and verify token (during decoding), the JWT package requires a JWK store. The JWK
store is responsible for providing the private and public keys used in the signing and verification process. These implementations
of the JWK stores are provided:

1. ```FileJwkStore```
2. ```RemoteJwkStore```

For testing, we also provide

1. ```SingleJwkStore``` - A JWK store that contains a randomly generated private key for the specified algorithm.
2. ```StaticJwkStore``` - A JWK store that contains a static list of randomly generated private key for the specified algorithm.

### FileJwkStore
This store is used to load JWKs from pem files. The file location is specified using configuration properties. This is the
default store used in the authorization server configuration.

```yaml
security:
  keys:
    my-key-name:
      id: my-key-id
      format: pem
      file: my-key-file.pem
```

The FileJwkStore will load all the keys under the `security.keys` property. The pem file under each key name can contain
one or more keys. Key name is a way to categorize the key by usage. For example, you may want to have a different set of keys
for signing and encryption.

If the pem file contains one key. The key id of the key will equal to the name of the key. In this example, if `my-key-file.pem` 
contains only one key. That key's key id will be `my-key-name`.

If the pem file contains multiple key. The key id of the key will either be based on the `id` property if it's provided. Or it will
be generated based on elements of the public key. In this example, since `id` is provided, the key id will be `my-key-id-1` and `my-key-id-2` etc.

#### Key Rotation
The FileJwkStore supports key rotation. If the pem file contains multiple keys, the `LoadByName` will return the current key for that name.
After `Rotate` is called, the current key will be moved to the next key in the pem file.

#### Supported Key Types
The FileJwkStore supports the following key types:

- RSA: PKCS8 unencrypted format. Tradition encrypted and unencrypted format. 
- ECDSA: PKCS8 unencrypted format. Tradition encrypted and unencrypted format.
- ED25519: PKCS8 unencrypted format. 
- HMAC: Custom unencrypted format.

See the [testdata](testdata/README.md) directory for examples of pem files and how to generate them using `openssl`.

#### Use HMAC Key with Caution
HMAC key is a symmetric key. This file store supports HMAC key. However, it should be used with caution. By default, the HMAC key
is included in the jwks endpoint which is by default public. If you want to use HMAC key, you should secure the jwks endpoint in 
your application, or encrypt the jwks content by providing your own jwks implementation instead of the default one.

### RemoteJwkStore
This store is used to load JWKs from a remote endpoint. It's usually used when your application needs to verify the jwt
signature issued from an authorization server that publishes its public keys through its jwks endpoint.