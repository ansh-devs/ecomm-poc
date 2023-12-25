package order

import (
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/hashicorp/consul/api"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func (s *OrderService) RegisterService(addr *string) {
	client := &api.AgentServiceCheck{
		CheckID:                        "status_alive",
		TTL:                            "5s",
		Name:                           "ORDER_SERVICE",
		TLSSkipVerify:                  true,
		DeregisterCriticalServiceAfter: "2s",
	}

	port, err := strconv.Atoi(strings.Trim(*addr, ":"))
	if err != nil {
		_ = level.Error(s.logger).Log("err", err)
	}
	regis := &api.AgentServiceRegistration{
		ID:      "_instance_" + strconv.Itoa(rand.Int()),
		Name:    "ORDER_SERVICE",
		Tags:    []string{"order"},
		Port:    port,
		Address: s.getLocalIP().String(),
		Meta:    map[string]string{"registered_at": time.Now().String()},
		Check:   client,
	}
	//if err := s; err != nil {
	if err := s.consulClient.Agent().ServiceRegister(regis); err != nil {
		_ = level.Error(s.logger).Log("err", err)
	}
}

func (s *OrderService) UpdateHealthStatus() {
	ticker := time.NewTicker(time.Second * 3)
	for {
		if err := s.consulClient.Agent().UpdateTTL("status_alive", "[MESSAGE]: working as intended", api.HealthPassing); err != nil {
			_ = level.Error(s.logger).Log("err", err)
		}
		<-ticker.C
	}
}

func (s *OrderService) getLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("%s\n", err.Error())
		}
	}(conn)

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}
