partkyle/docker-dns
==========

DNS server that returns the internal docker ip addresses for running containers. There is no caching and no TTL. It is mainly meant as a development tool.

Container Requirements
============

- Connection to a docker server
  - via the DOCKER_HOST https server (requires DOCKER_CERT_HOME and certs available)
  - By mounting the docker socket over /var/run/docker.sock

Example Docker Container Run
==========

```
docker run --name dns -p 53:53/udp -d -v /var/run/docker.sock:/var/run/docker.sock partkyle/docker-dns
```

OSX Tips
========

If you enable routing to your DOCKER_HOST vm, you will be able to access docker IPs locally.

```
# assuming the docker ip is in the 172.17 range
sudo route -n add 172.17.0.0/16 $(boot2docker ip)
```

You can then add an entry to your /etc/resolvers:

```
sudo mkdir -p /etc/resolvers
echo "nameserver $(boot2docker ip)" | sudo tee /etc/resolver/docker
```

You should now have a file with at `/etc/resolver/docker` with the contents, with your IP of course.

```
nameserver 192.168.99.102
```

You can then run a DNS server on the boot2docker vm, binding to port 53/udp (for DNS) and mounting the docker socket.

```
docker run --name dns -p 53:53/udp -d -v /var/run/docker.sock:/var/run/docker.sock partkyle/docker-dns
```

You should now be able to resolve dns for anything with a ".docker" domain.

Example:

```
$ docker run -d --name redis redis
b9558812882ccf15119e92c853264ec8d0fb68697d6d4c2a21266f3d7349e0c1
$ ping redis.docker
PING redis.docker (172.17.6.92): 56 data bytes
64 bytes from 172.17.6.92: icmp_seq=0 ttl=63 time=2.337 ms
$ telnet redis.docker 6379
Trying 172.17.6.92...
Connected to redis.docker.
Escape character is '^]'.
ping
+PONG
```


It also supports rDNS lookups (though not efficiently at the moment)

```
$ dig @$(machine ip) -x $(docker inspect -f '{{.NetworkSettings.IPAddress}}' redis)

; <<>> DiG 9.8.3-P1 <<>> @192.168.99.102 -x 172.17.6.92
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 43075
;; flags: qr rd; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0
;; WARNING: recursion requested but not available

;; QUESTION SECTION:
;92.6.17.172.in-addr.arpa. IN  PTR

;; ANSWER SECTION:
92.6.17.172.in-addr.arpa. 0  IN  PTR redis.docker.

;; Query time: 39 msec
;; SERVER: 192.168.99.102#53(192.168.99.102)
;; WHEN: Thu Apr  2 15:43:10 2015
;; MSG SIZE  rcvd: 94
```
