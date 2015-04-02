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
	for _, question := range r.Question {
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

	certPath := os.Getenv("DOCKER_CERT_PATH")
	client, err := docker.NewTLSClient(os.Getenv("DOCKER_HOST"), path.Join(certPath, "cert.pem"), path.Join(certPath, "key.pem"), path.Join(certPath, "ca.pem"))
	if err != nil {
		log.Fatal(err)
	}

	handler := Handler{docker: client}

	server := dns.Server{}
	server.Net = *network
	server.Addr = *addr

	dns.Handle(".", &handler)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
