#!/bin/zsh

TMP_DIR=".tmp"
mkdir -p $TMP_DIR
rm -f $TMP_DIR/*

for i in {1..3}
do
  echo "-----BEGIN HMAC KEY-----" > $TMP_DIR/hmac-256-$i.pem
  openssl rand -base64 32 >> $TMP_DIR/hmac-256-$i.pem # we could also encrypt this with aes using openssl-enc but we would need to write our own DEK info header
  echo "-----END HMAC KEY-----" >> $TMP_DIR/hmac-256-$i.pem
done

echo "Merging PEM blocks..."
cat `find $TMP_DIR -type f -name 'hmac-256-*.pem' | sort` > hmac-256.pem
