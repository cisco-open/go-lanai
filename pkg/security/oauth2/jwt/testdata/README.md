# Notes

Test PEM Blocks are generated using LibreSSL 3.9.2.

Each PEM file have 3 blocks for key rotation testing. Use "Executable Shell" under each section to put everything together

## RSA

Executable Shell: [rsa-2048.sh](rsa-2048.sh)

Non-Encrypted:
```shell
openssl genrsa 2048 > rsa-2048-priv-key.pem
openssl rsa -in rsa-2048-priv-key.pem -pubout > rsa-2048-pub-key.pem
```

Encrypted:
```shell
openssl rsa -passout file:passwd.txt -in rsa-2048-priv-key.pem -out rsa-2048-priv-key-aes256.pem -aes-256-cbc
openssl genrsa -passout pass:WrongPass -out rsa-2048-priv-key-aes256-bad.pem -aes256 2048
```

Certificate:
```shell
openssl req -new -sha256 -key rsa-2048-priv-key.pem -out rsa-2048.csr -config ca.cnf
openssl req -x509 -sha256 -days 36500 -key rsa-2048-priv-key.pem -in rsa-2048.csr -out rsa-2048.crt
```

## ECDSA

> **Note**: Only NIST P-224, P-256, P-384, and P-521 are supported by Go. See [crypt/elliptic](https://pkg.go.dev/crypto/elliptic)

Executable Shell: 
- [ec-p256.sh](ec-p256.sh)
- [ec-p384.sh](ec-p384.sh)
- [ec-p521.sh](ec-p521.sh)

Non-Encrypted:
```shell
openssl ecparam -name prime256v1 -genkey -noout -out ec-p256-priv-key.pem
openssl ec -in ec-p256-priv-key.pem -pubout -out ec-p256-pub-key.pem

openssl ecparam -name secp384r1 -genkey -noout -out ec-p384-priv-key.pem
openssl ec -in ec-p384-priv-key.pem -pubout -out ec-p384-pub-key.pem

openssl ecparam -name secp521r1 -genkey -noout -out ec-p521-priv-key.pem
openssl ec -in ec-p521-priv-key.pem -pubout -out ec-p521-pub-key.pem
```

Encrypted:
```shell
openssl ec -passout file:passwd.txt -in ec-p256-priv-key.pem -out ec-p256-priv-key-aes256.pem -aes-256-cbc
openssl ec -passout pass:WrongPass -in ec-p256-priv-key.pem -out ec-p256-priv-key-aes256-bad.pem -aes-256-cbc

openssl ec -passout file:passwd.txt -in ec-p384-priv-key.pem -out ec-p384-priv-key-aes256.pem -aes-256-cbc
openssl ec -passout pass:WrongPass -in ec-p384-priv-key.pem -out ec-p384-priv-key-aes256-bad.pem -aes-256-cbc

openssl ec -passout file:passwd.txt -in ec-p521-priv-key.pem -out ec-p521-priv-key-aes256.pem -aes-256-cbc
openssl ec -passout pass:WrongPass -in ec-p521-priv-key.pem -out ec-p521-priv-key-aes256-bad.pem -aes-256-cbc
```

Certificate:
```shell
openssl req -new -sha256 -key ec-p256-priv-key.pem -out ec-p256.csr -config ca.cnf
openssl req -x509 -sha256 -days 36500 -key ec-p256-priv-key.pem -in ec-p256.csr -out ec-p256.crt

openssl req -new -sha256 -key ec-p384-priv-key.pem -out ec-p384.csr -config ca.cnf
openssl req -x509 -sha256 -days 36500 -key ec-p384-priv-key.pem -in ec-p384.csr -out ec-p384.crt

openssl req -new -sha256 -key ec-p521-priv-key.pem -out ec-p521.csr -config ca.cnf
openssl req -x509 -sha256 -days 36500 -key ec-p521-priv-key.pem -in ec-p521.csr -out ec-p521.crt
```

## MAC Secret

TBD

## ED25519

Executable Shell: [ed25519.sh](ed25519.sh)

Non-Encrypted:
```shell
openssl genpkey -algorithm ed25519 -out ed25519-priv-key.pem
openssl pkey -in ed25519-priv-key.pem -pubout -out ed25519-pub-key.pem
```

Encrypted:

> **Note**: LibreSSL 3.9.2's `pkey` command doesn't output encrypted private key using "traditional" format, i.e. no "DEK-Info" header.
> Therefore, we don't test these cases. Following  command is only for reference.

```shell
openssl pkey -passout file:passwd.txt -in ed25519-priv-key.pem -out ed25519-priv-key-aes256.pem -aes-256-cbc
```

Certificate:
```shell
openssl req -new -sha256 -key ed25519-priv-key.pem -out ed25519.csr -config ca.cnf
openssl req -x509 -sha256 -days 36500 -key ed25519-priv-key.pem -in ed25519.csr -out ed25519.crt
```

## JWK & JWT

JWK JSON files are generated using [Online JWK Generator](https://jwkset.com/generate) with the first block of corresponding public key files.

JWT are generated using [Online JWT Builder](https://dinochiesa.github.io/jwt/) with the first block of corresponding private key files.