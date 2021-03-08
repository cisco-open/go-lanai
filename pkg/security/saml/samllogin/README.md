1. Create saml private key and cert using the following command

openssl req -x509 -newkey rsa:2048 -keyout saml_test.key -out saml_test.cert -days 365 -nodes -subj "/CN=myservice.example.com"

use -passout option to add passphrase to the private key