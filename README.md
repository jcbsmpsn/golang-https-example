# TLS Connection Options in Golang

This is not an official Google product

Golang sample code for a minimal HTTPS client and server that demos:

* a server certificate that satisfies SAN requirements.
* a client that trusts a specific certificate.
* a server that authenticates the client based on the client certificate used
  in connection negotiation.

## Generating Keys

```sh
openssl req -x509 -nodes -newkey rsa:2048 -keyout server.key -out server.crt -days 3650
```

## Running the Server

```sh
go run https_server.go
```

## Running the Client

```sh
go run https_client.go
```

## Error/Solution

**Error:** `x509: certificate signed by unknown authority`

**Solution:** The certificate served by https_server is self signed. This message
means that the Go lang https library can't find a way to trust the certificate
the server is responding with.

