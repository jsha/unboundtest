# Unbound test server

This is a little HTTP server that makes it easy to test DNS lookups without
running your own Unbound instance, and get detailed logs. Useful for debugging
DNS issues with Let's Encrypt. See index.html for more details.

To run locally:

```
go run unboundtest.go
```

Then visit http://localhost:1232/.

Alternately:

```
docker build . --tag unboundtest
docker run unboundtest
```

Then use `docker ps` and `docker inspect` to find the IP address of the
unboundtest container, and visit that IP address on port 1232.

## CLI
```
Usage of unboundtest:
  -listen string
        The address on which to listen for incoming Web requests (default ":1232")
  -unboundAddress string
        The address the unbound.conf instructs Unbound to listen on (default "127.0.0.1:1053")
  -unboundConfig string
        The path to the unbound.conf file (default "unbound.conf")
  -unboundExec string
        The path to the unbound executable (default "unbound")
  -index string
        The path to the index.html (default "index.html")
```

## Deploying

This service runs at https://unboundtest.com/, on fly.io. It uses a Docker image
automatically built in GitHub Actions and pushed to the the GitHub Container
Registry. The build is kicked off by pushes to the `main` branch, and pushes
the image to a `latest` tag. See .github/workflows.

Once a set of changes is pushed and the latest image is built, if you have the
correct permissions (to jsha's Fly account), run `flyctl deploy` to redeploy the
service.
