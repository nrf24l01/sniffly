package snifpacket

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/pcap"
)

func ReceivePackets(handle *pcap.Handle, iface string, packets chan *SnifPacket, wg *sync.WaitGroup) {
    defer wg.Done()
    // Close packets channel when this sender exits so receivers can finish cleanly.
    // ReceivePackets is the only sender on the channel, so it's safe to close it here.
    defer close(packets)

    packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

    localIPs, localMAC, err := GetLocalAddrs(iface)
    filterEnabled := true
    if err != nil {
        log.Printf("failed to get local addresses for interface %s: %v; outgoing filtering disabled", iface, err)
        filterEnabled = false
    }

    // Diagnostics & counters
    var received uint64
    var dropped uint64
    pktCh := packetSource.Packets()
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

LOOP:
    for {
        select {
        case packet, ok := <-pktCh:
            if !ok {
                // packet source closed
                break LOOP
            }

            sp, err := ProcessPacket(packet)
            if err != nil {
                continue
            }

            if filterEnabled {
                if _, ok := localIPs[sp.SrcIP]; !ok && sp.SrcMAC != localMAC {
                    continue
                }
            }

            // Try to send without blocking forever. If consumers are slow, drop
            // packets and log periodically rather than blocking the capture.
            select {
            case packets <- sp:
                atomic.AddUint64(&received, 1)
            default:
                // channel full, drop packet
                atomic.AddUint64(&dropped, 1)
                if atomic.LoadUint64(&dropped)%100 == 0 {
                    log.Printf("packets channel full, dropped=%d, received=%d, len(packets)=%d", atomic.LoadUint64(&dropped), atomic.LoadUint64(&received), len(packets))
                }
            }

        case <-ticker.C:
            // periodic status
            log.Printf("capture status: received=%d dropped=%d queue_len=%d", atomic.LoadUint64(&received), atomic.LoadUint64(&dropped), len(packets))
        }
    }
    log.Printf("Packet receiving goroutine for interface %s exiting", iface)
}