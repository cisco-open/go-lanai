#!/bin/zsh

TMP_DIR=".tmp"
mkdir -p $TMP_DIR
rm -f $TMP_DIR/*

for i in {1..3}
do
  echo "Generating Private/Public Key Pair - $i..."
  # new keys
  openssl genpkey -algorithm ed25519 -out $TMP_DIR/ed25519-priv-key-$i.pem
  openssl pkey -in $TMP_DIR/ed25519-priv-key-$i.pem -pubout -out $TMP_DIR/ed25519-pub-key-$i.pem
  # encrypted
  openssl pkey -passout file:passwd.txt -in $TMP_DIR/ed25519-priv-key-$i.pem -out $TMP_DIR/ed25519-priv-key-aes256-$i.pem -aes-256-cbc
  # cert
  openssl req -new -sha256 -key $TMP_DIR/ed25519-priv-key-$i.pem -out $TMP_DIR/ed25519-$i.csr -config ca.cnf
  openssl req -x509 -sha256 -days 36500 -key $TMP_DIR/ed25519-priv-key-$i.pem -in $TMP_DIR/ed25519-$i.csr -out $TMP_DIR/ed25519-$i.crt
done

# multi-block PEM
echo "Merging PEM blocks..."
cat `find $TMP_DIR -type f -name 'ed25519-priv-key-*.pem' -a ! -name '*ed25519-priv-key-aes256-*.pem'` > ed25519-priv-key.pem
cat `find $TMP_DIR -type f -name 'ed25519-priv-key-aes256-*.pem'` > ed25519-priv-key-aes256.pem
cat `find $TMP_DIR -type f -name 'ed25519-pub-key-*.pem'` > ed25519-pub-key.pem
cat `find $TMP_DIR -type f -name 'ed25519-*.crt'` > ed25519-cert.pem

# wrong password
echo "Finalizing..."
openssl pkey -passout pass:WrongPass -in $TMP_DIR/ed25519-priv-key-1.pem -out ed25519-priv-key-aes256-bad.pem -aes-256-cbc