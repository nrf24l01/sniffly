package snifpacket

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopacket/gopacket"
)

func ReceivePackets(packetSource *gopacket.PacketSource, iface string, packets chan *SnifPacket, wg *sync.WaitGroup) {
    defer wg.Done()
    defer close(packets)

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

            select {
            case packets <- sp:
                atomic.AddUint64(&received, 1)
            default:
                // channel full, drop packet
                atomic.AddUint64(&dropped, 1)
                if atomic.LoadUint64(&dropped)%1000 == 0 {
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