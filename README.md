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

**Solution:** The certificate served by https_server is self signed. This
message means that the Go lang https library can't find a way to trust the
certificate the server is responding with.

There are two possible solutions.

1. Disable the client side certificate verification. This solution has the
   advantage of expediency, but the disadvantage of making your client code
   susceptible to man in the middle attacks.

    ```diff
    @@ -11,7 +11,9 @@ import (
     func main() {
            client := &http.Client{
                    Transport: &http.Transport{
    -                       TLSClientConfig: &tls.Config{},
    +                       TLSClientConfig: &tls.Config{
    +                               InsecureSkipVerify: true,
    +                       },
                    },
            }
    ```

    Add the server certificate to the list of certificate authorities that the
    client trusts.

2. Add the server certificate to the list of certificate authorities trusted by
   the client.

    ```diff
    @@ -9,9 +10,18 @@ import (
     )

     func main() {
    +       caCert, err := ioutil.ReadFile("server.crt")
    +       if err != nil {
    +               log.Fatal(err)
    +       }
    +       caCertPool := x509.NewCertPool()
    +       caCertPool.AppendCertsFromPEM(caCert)
    +
            client := &http.Client{
                    Transport: &http.Transport{
    -                       TLSClientConfig: &tls.Config{},
    +                       TLSClientConfig: &tls.Config{
    +                               RootCAs: caCertPool,
    +                       },
                    },
            }
    ```

