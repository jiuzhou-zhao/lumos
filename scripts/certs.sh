#!/bin/bash
#
domain="ymipro.com"

#
SOURCE="$0"
while [ -h "$SOURCE"  ]; do # resolve $SOURCE until the file is no longer a symlink
    DIR="$( cd -P "$( dirname "$SOURCE"  )" && pwd  )"
    SOURCE="$(readlink "$SOURCE")"
    [[ $SOURCE != /*  ]] && SOURCE="$DIR/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
DIR="$( cd -P "$( dirname "$SOURCE"  )" && pwd  )"
cd "${DIR}/../" || exit

#
rm -rf certs || true
mkdir certs
cd certs || exit


# ca
openssl genrsa -out ca.key 4096
openssl req -config "${DIR}/server.conf" -new -x509 -sha256 -key ca.key -days 3650 -out ca.crt -subj "//CN=${domain}"

# server
openssl genrsa -out server.key 2048
openssl req -new -sha256 -key server.key -subj "/CN=server.${domain}" -out server.csr
openssl x509 -req -extfile "${DIR}/server.conf" -extensions v3_req -sha256 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 3650

# client
openssl genrsa -out client.key 2048
openssl req -new -sha256 -key client.key -subj "/CN=client.${domain}" -out client.csr
openssl x509 -req -extfile "${DIR}/server.conf" -extensions v3_req -sha256 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt -days 3650


# proxy-server
openssl genrsa -out proxy-server.key 2048
openssl req -new -sha256 -key proxy-server.key -subj "/CN=proxy-server.${domain}" -out proxy-server.csr
openssl x509 -req -extfile "${DIR}/server.conf" -extensions v3_req -sha256 -in proxy-server.csr -CA server.crt -CAkey server.key -CAcreateserial -out proxy-server.crt -days 3650

# proxy-client
openssl genrsa -out proxy-client.key 2048
openssl req -new -sha256 -key proxy-client.key -subj "/CN=proxy-client.${domain}" -out proxy-client.csr
openssl x509 -req -extfile "${DIR}/server.conf" -extensions v3_req -sha256 -in proxy-client.csr -CA client.crt -CAkey client.key -CAcreateserial -out proxy-client.crt -days 3650
