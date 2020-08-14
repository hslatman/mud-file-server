#!/bin/sh

# Generate self signed root CA cert
openssl req -nodes -x509 -newkey rsa:4096 -keyout ca.key -out ca.crt -subj "/C=NL/ST=Overijssel/L=Enschede/O=ExampleSign/OU=root/CN=ExampleSign"

# Generate server cert to be signed
openssl req -nodes -newkey rsa:4096 -keyout server.key -out server.csr -subj "/C=NL/ST=Overijssel/L=Enschede/O=ExampleSign/OU=server/CN=ExampleSign"

# Sign the server cert
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

# Create server PEM file
cat server.key server.crt > server.pem