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
