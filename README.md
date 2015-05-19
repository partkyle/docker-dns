partkyle/docker-dns
==========

DNS server that returns the internal docker ip addresses for running containers. There is no caching and no TTL. It is mainly meant as a development tool.

# Container Requirements


- Connection to a docker server
  - via the DOCKER_HOST https server (requires DOCKER_CERT_HOME and certs available)
  - By mounting the docker socket over /var/run/docker.sock

# Example Docker Container Run

```
docker run --name dns -p 53:53/udp -d -v /var/run/docker.sock:/var/run/docker.sock partkyle/docker-dns
```

# OSX Tips

If you enable routing to your DOCKER_HOST vm, you will be able to access docker IPs locally.

```
# assuming the docker ip is in the 172.17 range
sudo route -n add 172.17.0.0/16 $(boot2docker ip)
```

You can then add an entry to your /etc/resolver:

```
sudo mkdir -p /etc/resolver
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

# Alternative Setup (dnsmasq)

Install dnsmasq

```
# on OSX (follow the instructions to get it to start on boot, etc)
brew install dnsmasq
```

Add these entries to the `/usr/local/etc/dnsmasq.conf` file.

(I use google DNS, but you can place whatever you want here.)

```
server=8.8.8.8
server=8.8.4.4
server=/docker/<insert the DOCKER_HOST ip here>
```

You can then replace the dns entries in the /etc/resolv.conf to

```
nameserver 127.0.0.1
```

Note: On OSX, you may want to change these through the network settings. To do this from the command line:

```
# list all available network interfaces
$ sudo networksetup listallnetworkservices [2015-04-03 10:33:45]
An asterisk (*) denotes that a network service is disabled.
Bluetooth DUN
Display Ethernet
Wi-Fi
Bluetooth PAN
# select the service(s) you want to override
$ sudo networksetup -setdnsservers "Wi-Fi" 127.0.0.1
```

Dns should be configured appropriately, and dnsmasq will forward *.docker domains to your containers.

```
$ dig redis.docker

; <<>> DiG 9.8.3-P1 <<>> redis.docker
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 10433
;; flags: qr rd ra; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;redis.docker.			IN	A

;; ANSWER SECTION:
redis.docker.		0	IN	A	172.17.7.73

;; Query time: 13 msec
;; SERVER: 127.0.0.1#53(127.0.0.1)
;; WHEN: Fri Apr  3 10:36:17 2015
;; MSG SIZE  rcvd: 58

$ dig google.com

; <<>> DiG 9.8.3-P1 <<>> google.com
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 30684
;; flags: qr rd ra; QUERY: 1, ANSWER: 11, AUTHORITY: 0, ADDITIONAL: 0

;; QUESTION SECTION:
;google.com.			IN	A

;; ANSWER SECTION:
google.com.		83	IN	A	74.125.224.130
google.com.		83	IN	A	74.125.224.142
google.com.		83	IN	A	74.125.224.137
google.com.		83	IN	A	74.125.224.128
google.com.		83	IN	A	74.125.224.136
google.com.		83	IN	A	74.125.224.133
google.com.		83	IN	A	74.125.224.132
google.com.		83	IN	A	74.125.224.135
google.com.		83	IN	A	74.125.224.129
google.com.		83	IN	A	74.125.224.134
google.com.		83	IN	A	74.125.224.131

;; Query time: 35 msec
;; SERVER: 127.0.0.1#53(127.0.0.1)
;; WHEN: Fri Apr  3 10:36:33 2015
;; MSG SIZE  rcvd: 204
```
