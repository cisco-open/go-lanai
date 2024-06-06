#!/bin/zsh

TMP_DIR=".tmp"
mkdir -p $TMP_DIR
rm -f $TMP_DIR/*

for i in {1..3}
do
  echo "Generating Private/Public Key Pair - $i..."
  # new keys
  openssl ecparam -name secp384r1 -genkey -noout -out $TMP_DIR/ec-p384-priv-key-$i.pem
  openssl ec -in $TMP_DIR/ec-p384-priv-key-$i.pem -pubout -out $TMP_DIR/ec-p384-pub-key-$i.pem
  # encrypted
  openssl ec -passout file:passwd.txt -in $TMP_DIR/ec-p384-priv-key-$i.pem -out $TMP_DIR/ec-p384-priv-key-aes256-$i.pem -aes-256-cbc
  # cert
  openssl req -new -sha256 -key $TMP_DIR/ec-p384-priv-key-$i.pem -out $TMP_DIR/ec-p384-$i.csr -config ca.cnf
  openssl req -x509 -sha256 -days 36500 -key $TMP_DIR/ec-p384-priv-key-$i.pem -in $TMP_DIR/ec-p384-$i.csr -out $TMP_DIR/ec-p384-$i.crt
done

# multi-block PEM
echo "Merging PEM blocks..."
cat `find $TMP_DIR -type f -name 'ec-p384-priv-key-*.pem' -a ! -name '*ec-p384-priv-key-aes256-*.pem'` > ec-p384-priv-key.pem
cat `find $TMP_DIR -type f -name 'ec-p384-priv-key-aes256-*.pem'` > ec-p384-priv-key-aes256.pem
cat `find $TMP_DIR -type f -name 'ec-p384-pub-key-*.pem'` > ec-p384-pub-key.pem
cat `find $TMP_DIR -type f -name 'ec-p384-*.crt'` > ec-p384-cert.pem

# wrong password
echo "Finalizing..."
openssl ec -passout pass:WrongPass -in $TMP_DIR/ec-p384-priv-key-1.pem -out ec-p384-priv-key-aes256-bad.pem -aes-256-cbc