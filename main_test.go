package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/miekg/dns"

	"testing"
)

type FakeDockerClient struct {
	inspectContainer func(string) (*docker.Container, error)
	listContainers   func(docker.ListContainersOptions) ([]docker.APIContainers, error)
}

func (f *FakeDockerClient) InspectContainer(s string) (*docker.Container, error) {
	return f.inspectContainer(s)
}

func (f *FakeDockerClient) ListContainers(o docker.ListContainersOptions) ([]docker.APIContainers, error) {
	return f.listContainers(o)
}

func setup(t *testing.T, domain string) (net.Listener, *FakeDockerClient, *dns.Server, *dns.Client) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	dockerClient := &FakeDockerClient{}
	handler := &Handler{domain: domain, docker: dockerClient}
	server := &dns.Server{Listener: listener, Handler: handler}
	go func() {
		err := server.ActivateAndServe()
		if err != nil {
			t.Fatal()
		}
	}()

	client := &dns.Client{}
	client.Net = "tcp"

	return listener, dockerClient, server, client
}

func TestDockerClientError(t *testing.T) {
	listener, dockerClient, _, client := setup(t, ".")

	dockerClient.inspectContainer = func(string) (*docker.Container, error) { return nil, fmt.Errorf("error") }

	msg := &dns.Msg{}
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.Question = []dns.Question{
		{Name: "api.docker.", Qclass: dns.ClassINET, Qtype: dns.TypeA},
	}

	_, _, err := client.Exchange(msg, listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	listener.Close()
}

func TestDockerClient_ARecords(t *testing.T) {

	container := "api"
	domain := ".docker."
	ip := "127.0.0.1"

	listener, dockerClient, _, client := setup(t, domain)

	dockerClient.inspectContainer = func(c string) (*docker.Container, error) {
		if c != container {
			return nil, fmt.Errorf("container does not exist; have %s, want %s:", c, container)
		}
		return &docker.Container{NetworkSettings: &docker.NetworkSettings{IPAddress: ip}}, nil
	}

	msg := &dns.Msg{}
	msg.Question = []dns.Question{
		{Name: container + domain, Qclass: dns.ClassINET, Qtype: dns.TypeA},
	}

	reply, _, err := client.Exchange(msg, listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	expectedReplies := 1
	if len(reply.Answer) != expectedReplies {
		t.Errorf("wrong number of replies; have %d, want %d", len(reply.Answer), expectedReplies)
	}

	for _, answer := range reply.Answer {
		if !strings.HasSuffix(answer.String(), ip) {
			t.Errorf("did not get expected ip in %q", answer.String())
		}
	}
}

func TestDockerClient_ARecords_NoResults(t *testing.T) {
	container := "api"
	domain := ".docker."
	ip := "127.0.0.1"

	listener, dockerClient, _, client := setup(t, domain)

	dockerClient.inspectContainer = func(c string) (*docker.Container, error) {
		if c != container {
			return nil, fmt.Errorf("container does not exist; have %s, want %s:", c, container)
		}
		return &docker.Container{NetworkSettings: &docker.NetworkSettings{IPAddress: ip}}, nil
	}

	msg := &dns.Msg{}
	msg.Question = []dns.Question{
		{Name: "imnothere" + domain, Qclass: dns.ClassINET, Qtype: dns.TypeA},
	}

	reply, _, err := client.Exchange(msg, listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	expectedReplies := 0
	if len(reply.Answer) != expectedReplies {
		t.Errorf("wrong number of replies; have %d, want %d", len(reply.Answer), expectedReplies)
	}
}

func TestDockerClient_MXRecords(t *testing.T) {
	container := "api"
	domain := ".docker."
	ip := "127.0.0.1"

	listener, dockerClient, _, client := setup(t, domain)

	dockerClient.inspectContainer = func(c string) (*docker.Container, error) {
		if c != container {
			return nil, fmt.Errorf("container does not exist; have %s, want %s:", c, container)
		}
		return &docker.Container{NetworkSettings: &docker.NetworkSettings{IPAddress: ip}}, nil
	}

	msg := &dns.Msg{}
	msg.Question = []dns.Question{
		{Name: container + domain, Qclass: dns.ClassINET, Qtype: dns.TypeMX},
	}

	reply, _, err := client.Exchange(msg, listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	expectedReplies := 1
	if len(reply.Answer) != expectedReplies {
		t.Errorf("wrong number of replies; have %d, want %d", len(reply.Answer), expectedReplies)
	}

	for _, answer := range reply.Answer {
		if !strings.HasSuffix(answer.String(), container+domain) {
			t.Errorf("did not get expected ip in %q", answer.String())
		}
	}
}

func TestDockerClient_MXRecords_NoResults(t *testing.T) {
	container := "api"
	domain := ".docker."
	ip := "127.0.0.1"

	listener, dockerClient, _, client := setup(t, domain)

	dockerClient.inspectContainer = func(c string) (*docker.Container, error) {
		if c != container {
			return nil, fmt.Errorf("container does not exist; have %s, want %s:", c, container)
		}
		return &docker.Container{NetworkSettings: &docker.NetworkSettings{IPAddress: ip}}, nil
	}

	msg := &dns.Msg{}
	msg.Question = []dns.Question{
		{Name: "imnothere" + domain, Qclass: dns.ClassINET, Qtype: dns.TypeMX},
	}

	reply, _, err := client.Exchange(msg, listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	expectedReplies := 0
	if len(reply.Answer) != expectedReplies {
		t.Errorf("wrong number of replies; have %d, want %d", len(reply.Answer), expectedReplies)
	}
}
