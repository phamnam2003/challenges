#!/bin/bash
set -e

echo "generating a self-signing certificate authority key..."
openssl genrsa -out ./tech/scylladb/tls/cadb.key 4096

echo "a certificate signing authority..."
openssl req -x509 -new -nodes -key ./tech/scylladb/tls/cadb.key -days 3650 -config ./tech/scylladb/tls/db.cfg -out ./tech/scylladb/tls/cadb.pem

echo "generate a private key for our certificate..."
openssl genrsa -out ./tech/scylladb/tls/db.key 4096

echo "signing request..."
openssl req -new -key ./tech/scylladb/tls/db.key -out ./tech/scylladb/tls/db.csr -config ./tech/scylladb/tls/db.cfg

echo "create and sign our certificate..."
openssl x509 -req -in ./tech/scylladb/tls/db.csr -CA ./tech/scylladb/tls/cadb.pem -CAkey ./tech/scylladb/tls/cadb.key -CAcreateserial -out ./tech/scylladb/tls/db.crt -days 365 -sha256
