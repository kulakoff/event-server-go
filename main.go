package main

import (
	"fmt"
	"github.com/kulakoff/event-server-go/internal/config"
	"github.com/kulakoff/event-server-go/internal/services"
	"log"
	"log/slog"
	"net"
)

func main() {
	slog.Info("app started")
	cfg, err := config.New("config.json")
	if err != nil {
		log.Fatalf("Error load config file: %v", err)
	}

	go startServer(cfg.Hw.Beward.Port, "Beward")
	go startServer(cfg.Hw.BewardDS.Port, "BewardDS")
	go startServer(cfg.Hw.Qtech.Port, "Qtech")

	select {}

}

func startServer(port int, panelType string) {
	addr := fmt.Sprintf(":%d", port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Error resolving UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Error listening UDP: %v", err)
	}
	defer conn.Close()

	log.Printf("Server %s running on port %d ", panelType, port)

	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		message := string(buffer[:n])
		log.Printf("HW: %s | %s", panelType, message)

		/**
		TODO:
			- implements service handlers by device type
		*/
		service := services.GetIntercomService(panelType)
		if err := service.ProcessSyslogMessage(message); err != nil {
			log.Printf("Error processing syslog message from hw %s: %v", panelType, err)
		}
	}
}
