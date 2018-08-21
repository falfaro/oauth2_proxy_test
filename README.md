# oauth2_proxy_test

This repository implements a very simple setup based on Docker compose to test the
integration between `oauth2_proxy` and the `dex` OpenID Connect server configured
as a mock. This OAuth2 environment is used to protect a simple Web page served by
`ngnix`.

To test:

```
$ docker-compose up --build
```

This will spawn three containers:

* `dex`: a mock OpenID Connect server based on CoreOS Dex which allows authenticating
         with a user named `admin` and password `password` or by using some example
         credentials. `dex` listens on `http://172.30.0.3:5556`.
* `authproxy`: The `oauth2_proxy` binary configured to integrate with `dex` and
               protecting a simple Web page served by the `upstream` container.
               `oauth2_proxy` listens on `http://172.30.0.4:4180` and is configured to
               forward traffic to `upstream`.
* `upstream`: The protected resource, which consists of a simple Web page served by
              NGNIX, listening on `http://172.30.0.5`.
* `firefox`: Firefox running inside a Docker container and exposed via VNC listening
             on `http://localhost:5800`.

Once the containers are running, point your browser to `http://localhost:5800` in order
to access Firefox running in the container. Point Firefox to the `authproxy` container,
`http://172.30.0.4:4180` to trigger the OAuth2 authentication before being able to
access the protected resource. Once the OAuth2 flow has been completed, browsing to
`http://172.30.0.4:4180` should show the protected resource without triggering any
authentication.

To run the automated test (written in Go), launch the Docker stack, then:

```
docker exec -it firefox bash
cd /go
go run test.go
```