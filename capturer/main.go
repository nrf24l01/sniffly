package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gopacket/gopacket/pcap"
	"github.com/joho/godotenv"
	"github.com/nrf24l01/sniffly/capturer/core"
	"github.com/nrf24l01/sniffly/capturer/grpc"
	"github.com/nrf24l01/sniffly/capturer/snifpacket"
)

func main() {
    // Try to load .env file in non-production environment
	if os.Getenv("PRODUCTION_ENV") != "true" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("failed to load .env: %v", err)
		}
	}

    // Load configuration from environment variables
    config := core.LoadConfigFromEnv()

    // Initialize packet channel
    packets := make(chan *snifpacket.SnifPacket, 100)

    // Start gRPC connection and streaming in a separate goroutine
    client, err := grpc.ConnectPacketGatewayClient(config)
    if err != nil {
        log.Fatalf("Failed to connect to gRPC server: %v", err)
    }

    // Open device for packet capturing
    handle, err := pcap.OpenLive(config.Interface, 1600, true, pcap.BlockForever)
    if err != nil {
        log.Fatal(err)
    }
    defer handle.Close()

    // Gorutines
    var wg sync.WaitGroup
    fmt.Printf("Starting packet capture on interface: %s to target %s\n", config.Interface, config.ServerAddress)

    // Start packet processing loop
    go snifpacket.ReceivePackets(handle, packets, &wg)
    wg.Add(1)
    go grpc.StreamPackets(client, config, packets, &wg)
    wg.Add(1)

    // On exit
    wg.Wait()
    fmt.Printf("Exiting...")
}
