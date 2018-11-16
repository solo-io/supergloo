#!/usr/bin/env bash

set -ex

openssl ecparam -name secp384r1 -out ec.param
openssl req -new -x509 -nodes -newkey ec:ec.param -keyout root-key.pem -out root-cert.pem -days 365 -subj /C=US/ST=Massachusetts/L=Cambridge/O=Org/CN=www.example.com
rm ec.param

