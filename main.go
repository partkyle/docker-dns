package main

import (
	"flag"
	"net"
	"os"
	"path"
	"strings"

	docker "github.com/fsouza/go-dockerclient"

	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

var (
	network = flag.String("net", "udp", "network type (tcp/udp)")
	addr    = flag.String("addr", ":53", "addr to bind to")
)

type Handler struct {
	docker *docker.Client
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	reply := &dns.Msg{}
	reply.SetReply(r)

	answer := make([]dns.RR, 0)
lookup:
	for _, question := range r.Question {
		log.WithField("type", question.Qtype).WithField("name", question.Name).Info("resolving dns")

		switch question.Qtype {
		case dns.TypeA:
			container, err := h.docker.InspectContainer(strings.TrimSuffix(question.Name, ".docker."))
			if err != nil {
				log.Error(err)
				continue
			}
			a := &dns.A{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeA,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				A: net.ParseIP(container.NetworkSettings.IPAddress),
			}

			answer = append(answer, a)
		case dns.TypeMX:
			mx := &dns.MX{
				Hdr: dns.RR_Header{
					Name:   question.Name,
					Rrtype: dns.TypeMX,
					Class:  dns.ClassINET,
					Ttl:    0,
				},
				Preference: 10,
				Mx:         question.Name,
			}

			answer = append(answer, mx)
		case dns.TypePTR:
			containers, err := h.docker.ListContainers(docker.ListContainersOptions{})
			if err != nil {
				log.Error(err)
				continue
			}
			// these look like: 73.6.17.172.in-addr.arpa.
			reverseIP := strings.TrimSuffix(question.Name, ".in-addr.arpa.")
			reverseParts := strings.Split(reverseIP, ".")

			parts := make([]string, len(reverseParts))
			for i := range parts {
				parts[i] = reverseParts[len(reverseParts)-i-1]
			}

			ip := strings.Join(parts, ".")
			log.WithField("ip", ip).Info("rDNS lookup")

			for _, c := range containers {
				container, err := h.docker.InspectContainer(c.ID)
				if err != nil {
					log.Error(err)
					continue
				}

				if container.NetworkSettings.IPAddress == ip {
					ptr := &dns.PTR{
						Hdr: dns.RR_Header{
							Name:   question.Name,
							Rrtype: dns.TypePTR,
							Class:  dns.ClassINET,
							Ttl:    0,
						},
						Ptr: strings.TrimPrefix(container.Name, "/") + ".docker.",
					}
					answer = append(answer, ptr)

					continue lookup
				}
			}
		}
	}

	reply.Answer = answer

	err := w.WriteMsg(reply)
	if err != nil {
		log.Error(err)
	}
}

func main() {
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	var client *docker.Client
	if os.Getenv("DOCKER_HOST") != "" {
		var err error
		certPath := os.Getenv("DOCKER_CERT_PATH")
		client, err = docker.NewTLSClient(os.Getenv("DOCKER_HOST"), path.Join(certPath, "cert.pem"), path.Join(certPath, "key.pem"), path.Join(certPath, "ca.pem"))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		var err error
		client, err = docker.NewClient("unix:///var/run/docker.sock")
		if err != nil {
			log.Fatal(err)
		}
	}

	handler := Handler{docker: client}

	server := dns.Server{}
	server.Net = *network
	server.Addr = *addr

	dns.Handle(".", &handler)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
