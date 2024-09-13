package main

import (
	"fmt"
	"github.com/kulakoff/event-server-go/internal/config"
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

	//fmt.Printf("Bwward event server used port: %d\n", cfg.Hw.Beward.Port)
	//fmt.Printf("Bwward DS event server used port: %d\n", cfg.Hw.BewardDs.Port)
	//fmt.Printf("Qtech event server used port: %d\n", cfg.Hw.Qtech.Port)

	go startServer(cfg.Hw.Beward.Port, "Beward")
	go startServer(cfg.Hw.BewardDs.Port, "BewardDS")
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
		log.Printf("Received message: %s", message)

		/**
		TODO:
			- implements service handlers by device type
		*/
	}
}
