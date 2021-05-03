#!/usr/bin/env bash

openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:4096 -subj "/C=US/ST=CA/O=Solo.io" -keyout echo-ca.key -out echo-ca.crt
openssl req -out echo.csr -newkey rsa:2048 -nodes -keyout echo.key -config echo.cnf
openssl x509 -req -days 365 -CA echo-ca.crt -CAkey echo-ca.key -set_serial 0 -in echo.csr -out echo.crt -extfile echo.cnf -extensions san_reqext
