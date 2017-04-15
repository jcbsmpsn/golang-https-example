# TLS Connection Options in Golang

This is not an official Google product

Golang sample code for a minimal HTTPS client and server that demos:

* a server certificate that satisfies SAN requirements.
* a client that trusts a specific certificate.
* a server that authenticates the client based on the client certificate used
  in connection negotiation.

## Generating Key and Self Signed Cert

```sh
openssl req \
    -x509 \
    -nodes \
    -newkey rsa:2048 \
    -keyout server.key \
    -out server.crt \
    -days 3650 \
    -subj "/C=GB/ST=London/L=London/O=Global Security/OU=IT Department/CN=*"
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

**Error:** `x509: invalid signature: parent certificate cannot sign this kind of certificate`

**Solution:** The wrong kind of server certificate was generated. The property
in the CA that signed the server certificate indicates that the signing
certificate is not a CA. Since this is a self signed server certificate, it
needs the signing permission to sign itself.

Using `openssl x509 -in server.crt -text -noout` to look at the details of the
server certificate reveals that it is missing the CA flag, which should look
like this:

```
            X509v3 Basic Constraints:
                CA:TRUE
```

**Error:** `x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs`

**Solution:** A SAN is a Subject Alternative Name, an x509 extension that
allows additional names to be specified as valid domains for the certficate.

Starting with Go 1.3, when connecting to a server via the IP address rather
than the hostname, the CN field in the server certificate is ignored by the
client golang libraries and names specified as SANs will be used instead.

Using `openssl x509 -in server.crt -text -noout` to look at the details of the
server certificate reveals that it is missing a SAN section, which should look
like this:

```
        X509v3 extensions:
            X509v3 Subject Alternative Name:
                IP Address:127.0.0.1
```

There are two possible solutions.

1. Use a name to connect to the server instead of an IP address. If the client
   connects with a name matching the certificate CN, a SAN is not required.

    ```diff
    -       resp, err := client.Get("https://127.0.0.1:8443")
    +       resp, err := client.Get("https://localhost:8443")
    ```

    Using `openssl x509 -in server.crt -text -noout` to look at the **Subject**
    line should show **CN=** matching the name of the server. `localhost` or
    `*` will work.

    ```
            Subject: CN=*
    ```

2. Add a SAN to the certificate with the IP address of the server.

    To add a SAN to a certificate, there is multiple steps required, that will
    generate a separate CA and use that to sign the server certificate signing
    request.

    ```sh
    openssl req \
        -newkey rsa:2048 \
        -nodes \
        -days 3650 \
        -x509 \
        -keyout ca.key \
        -out ca.crt \
        -subj "/CN=*"
    openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout server.key \
        -out server.csr \
        -subj "/C=GB/ST=London/L=London/O=Global Security/OU=IT Department/CN=*"
    openssl x509 \
        -req \
        -days 365 \
        -sha256 \
        -in server.csr \
        -CA ca.crt \
        -CAkey ca.key \
        -CAcreateserial \
        -out server.crt \
        -extfile <(echo subjectAltName = IP:127.0.0.1)
    ```

**Error:** `tls: client didn't provide a certificate`

**Solution:** When the server code has the option set to authenticate client
connections using the client certificate, the server will drop connections from
clients using certs that are untrusted.

```diff
-       cfg := &tls.Config{}
+       cfg := &tls.Config{
+               ClientAuth: tls.RequireAndVerifyClientCert,
+       }
```

Generate a client certificate to use:

```sh
openssl req \
    -x509 \
    -nodes \
    -newkey rsa:2048 \
    -keyout client.key \
    -out client.crt \
    -days 3650 \
    -subj "/C=GB/ST=London/L=London/O=Global Security/OU=IT Department/CN=*"
```

The client has to be configured to send a certificate with connection attempts:

```diff
+       cert, err := tls.LoadX509KeyPair("client.crt", "client.key")
+       if err != nil {
+               log.Fatal(err)
+       }
+
        client := &http.Client{
                Transport: &http.Transport{
                        TLSClientConfig: &tls.Config{
                                RootCAs:      caCertPool,
+                               Certificates: []tls.Certificate{cert},
                        },
                },
        }
```

And the server has to be configured to trust the client certificate:

```diff
+       caCert, err := ioutil.ReadFile("client.crt")
+       if err != nil {
+               log.Fatal(err)
+       }
+       caCertPool := x509.NewCertPool()
+       caCertPool.AppendCertsFromPEM(caCert)
        cfg := &tls.Config{
                ClientAuth: tls.RequireAndVerifyClientCert,
+               ClientCAs:  caCertPool,
        }
```

