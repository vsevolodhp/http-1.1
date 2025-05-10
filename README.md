# Useless HTTP/1.1
I wanted to try to build HTTP server directly on top of TCP -- just for fun. No reason other than exploring how HTTP actually works.

Before doing that I read a couple of RFCs:
- [RFC 2616 -- Hypertext Transfer Protocol](https://datatracker.ietf.org/doc/html/rfc2616) (now obsolete).
- [RFC 7230 -- Hypertext Transfer Protocol: Message Syntax and Routing](https://datatracker.ietf.org/doc/html/rfc7230).

It was fun experience🙂

Expect no safety, no performance, and no TLS.

## Try it
```bash
go run main.go
```

In another terminal tab:
```bash
curl http://localhost:8080/ping -v
```
Beware: `curl` will terminate the connection. I used `nc` to test persistent connections:
```
nc localhost 8080
```
Then manually typed out HTTP message, for example:
```plaintext
GET /ping HTTP/1.1
HOST: localhost
Connection: close
```
