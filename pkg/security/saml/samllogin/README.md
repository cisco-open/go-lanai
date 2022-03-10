1. Create saml private key and cert using the following command

openssl genrsa -out saml.key -aes256 1024

openssl req -key saml.key -new -x509 -days 36500 -out saml.crt
