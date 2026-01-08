package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/afpacket"
	"github.com/gopacket/gopacket/layers"
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
    packets := make(chan *snifpacket.SnifPacket, 1000)

    // Start gRPC connection and streaming in a separate goroutine
    client, err := grpc.ConnectPacketGatewayClient(config)
    if err != nil {
        log.Fatalf("Failed to connect to gRPC server: %v", err)
    }

    // Open device for packet capturing via AF_PACKET (Linux)
    tp, err := afpacket.NewTPacket(
        afpacket.OptInterface(config.Interface),
        afpacket.OptFrameSize(65536),
        afpacket.OptBlockSize(1024*1024),
        afpacket.OptNumBlocks(32),
        afpacket.OptPollTimeout(500*time.Millisecond),
    )
    if err != nil {
        log.Fatalf("failed to open AF_PACKET on %s: %v", config.Interface, err)
    }
    defer tp.Close()

    packetSource := gopacket.NewPacketSource(tp, layers.LinkTypeEthernet)
    packetSource.NoCopy = true

    // Goroutines
    var wg sync.WaitGroup
    fmt.Printf("Starting packet capture on interface: %s to target %s\n", config.Interface, config.ServerAddress)

    // Start packet processing
    wg.Add(1)
    go snifpacket.ReceivePackets(packetSource, config.Interface, packets, &wg)

    wg.Add(1)
    go func() {
        if err := grpc.StreamPackets(client, config, packets, &wg); err != nil {
            log.Printf("StreamPackets exited with error: %v", err)
        } else {
            log.Printf("StreamPackets exited normally")
        }
    }()

    // On exit
    wg.Wait()
    fmt.Printf("Exiting...")
}
