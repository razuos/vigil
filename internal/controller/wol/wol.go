package wol

import (
	"fmt"
	"net"
)

// SendMagicPacket sends a WOL packet to the specified MAC address.
func SendMagicPacket(macAddr string) error {
	hwAddr, err := net.ParseMAC(macAddr)
	if err != nil {
		return fmt.Errorf("invalid MAC address: %w", err)
	}

	// Constuct Magic Packet: 6x FF followed by 16x MAC
	packet := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	for i := 0; i < 16; i++ {
		packet = append(packet, hwAddr...)
	}

	// Send via UDP to broadcast
	conn, err := net.Dial("udp", "255.255.255.255:9")
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}
	defer conn.Close()

	if _, err := conn.Write(packet); err != nil {
		return fmt.Errorf("failed to write packet: %w", err)
	}

	return nil
}
