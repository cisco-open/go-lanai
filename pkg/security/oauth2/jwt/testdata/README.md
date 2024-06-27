# Notes

Test PEM Blocks are generated using LibreSSL 3.9.2.

Each PEM file have 3 blocks for key rotation testing. Use "Executable Shell" under each section to put everything together

## RSA

Executable Shell: [rsa-2048.sh](rsa-2048.sh)

Non-Encrypted. The first command generates the private key in "traditional" format. The second command converts it to PKCS8 format. 
The third command generates the public key:
```shell
openssl genrsa 2048 -out rsa-2048-priv-key-trad.pem
openssl pkcs8 -topk8 -nocrypt -in rsa-2048-priv-key-trad.pem -out rsa-2048-priv-key.pem
openssl rsa -in rsa-2048-priv-key.pem -pubout > rsa-2048-pub-key.pem
```

Encrypted. Both command results in "traditional" format pem. The first command encrypts the private key generated in the
previous step. The second command generates a new private key with "WrongPass" as the passphrase. This is used to test 
the case where the passphrase is incorrect:
```shell
openssl rsa -passout file:passwd.txt -in $TMP_DIR/rsa-2048-priv-key-$i.pem -out $TMP_DIR/rsa-2048-priv-key-aes256-$i.pem -aes-256-cbc
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

Non-Encrypted. In each set of command, the first command is to generate the private key in "traditional" format. 
The second command is to convert it to PKCS8 format. The third command generates the public key:
```shell
openssl ecparam -name prime256v1 -genkey -noout -out ec-p256-priv-key-trad.pem
openssl pkcs8 -topk8 -nocrypt -in ec-p256-priv-key-trad.pem -out ec-p256-priv-key.pem
openssl ec -in ec-p256-priv-key.pem -pubout -out ec-p256-pub-key.pem

openssl ecparam -name secp384r1 -genkey -noout -out ec-p384-priv-key-trad.pem
openssl pkcs8 -topk8 -nocrypt -in ec-p384-priv-key-trad.pem -out ec-p384-priv-key.pem
openssl ec -in ec-p384-priv-key.pem -pubout -out ec-p384-pub-key.pem

openssl ecparam -name secp521r1 -genkey -noout -out ec-p521-priv-key-trad.pem
openssl pkcs8 -topk8 -nocrypt -in ec-p521-priv-key-trad.pem -out ec-p521-priv-key.pem
openssl ec -in ec-p521-priv-key.pem -pubout -out ec-p521-pub-key.pem
```

Encrypted. In each set of command, the first command generates an encrypted private key in traditional format. The second
command generates an encrypted private key using "WrongPass" as the passphrase. This is used to test the case where the
passphrase is incorrect:
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
Mac secret is a symmetric key. There is no official pem format or openssl command to generate it. The openssl rand 
command is used to generate a base64 encoded key with the corresponding bytes. We use a custom header to label the pem block.

Executable Shell:
- [hmac-256.sh](hmac-256.sh)
- [hmac-384.sh](hmac-384.sh)
- [hmac-512.sh](hmac-512.sh)

Non-Encrypted:
```shell
  echo "-----BEGIN HMAC KEY-----" > hmac-256.pem
  openssl rand -base64 32 >> hmac-256-$i.pem
  echo "-----END HMAC KEY-----" >> hmac-256.pem
```

Encrypted:
> **Note**: LibreSSL 3.9.2 does not have a command that encrypts a symmetric key. Therefore, we don't test these cases.

## ED25519

Executable Shell: [ed25519.sh](ed25519.sh)

Non-Encrypted, the first command generates the private key in pkcs8 format. The second command is used to convert to pkcs8 format.
However, since the file is already in pkcs8 format, this command is redundant. It's included for reference only.
The third command generates the public key:
```shell
openssl genpkey -algorithm ed25519 -out ed25519-priv-key-trad.pem
openssl pkcs8 -topk8 -nocrypt -in ed25519-priv-key-trad.pem -out ed25519-priv-key.pem
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