#!/bin/zsh

TMP_DIR=".tmp"
mkdir -p $TMP_DIR
rm -f $TMP_DIR/*

for i in {1..3}
do
  echo "Generating Private/Public Key Pair - $i..."
  # new keys
  openssl genrsa 2048 -out $TMP_DIR/rsa-2048-priv-key-trad-$i.pem
  openssl pkcs8 -topk8 -nocrypt -in $TMP_DIR/rsa-2048-priv-key-trad-$i.pem -out $TMP_DIR/rsa-2048-priv-key-$i.pem
  openssl rsa -in $TMP_DIR/rsa-2048-priv-key-$i.pem -pubout > $TMP_DIR/rsa-2048-pub-key-$i.pem
  # encrypted
  openssl rsa -passout file:passwd.txt -in $TMP_DIR/rsa-2048-priv-key-$i.pem -out $TMP_DIR/rsa-2048-priv-key-aes256-$i.pem -aes-256-cbc
  # cert
  openssl req -new -sha256 -key $TMP_DIR/rsa-2048-priv-key-$i.pem -out $TMP_DIR/rsa-2048-$i.csr -config ca.cnf
  openssl req -x509 -sha256 -days 36500 -key $TMP_DIR/rsa-2048-priv-key-$i.pem -in $TMP_DIR/rsa-2048-$i.csr -out $TMP_DIR/rsa-2048-$i.crt
done

# multi-block PEM
echo "Merging PEM blocks..."
cat `find $TMP_DIR -type f -name 'rsa-2048-priv-key-*.pem' -a ! -name '*rsa-2048-priv-key-aes256-*.pem' -a ! -name '*rsa-2048-priv-key-trad-*.pem' | sort` > rsa-2048-priv-key.pem
cat `find $TMP_DIR -type f -name 'rsa-2048-priv-key-trad-*.pem' | sort` > rsa-2048-priv-key-trad.pem
cat `find $TMP_DIR -type f -name 'rsa-2048-priv-key-aes256-*.pem' | sort` > rsa-2048-priv-key-aes256.pem
cat `find $TMP_DIR -type f -name 'rsa-2048-pub-key-*.pem' | sort` > rsa-2048-pub-key.pem
cat `find $TMP_DIR -type f -name 'rsa-2048-*.crt' | sort` > rsa-2048-cert.pem

# wrong password
echo "Finalizing..."
openssl genrsa -passout pass:WrongPass -out rsa-2048-priv-key-aes256-bad.pem -aes256 2048