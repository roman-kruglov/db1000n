package packetgen

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"golang.org/x/net/ipv4"
)

type PacketConfig struct {
	Ethernet EthernetPacketConfig
	IP       IPPacketConfig
	TCP      TCPPacketConfig
	Payload  []byte
}

func SendPacket(c PacketConfig, destinationHost string, destinationPort int) (int, error) {
	var (
		ipHeader   *ipv4.Header
		packetConn net.PacketConn
		rawConn    *ipv4.RawConn
		err        error
	)
	tcpPacket := buildTcpPacket(c.TCP)
	ipPacket := buildIpPacket(c.IP)
	if err = tcpPacket.SetNetworkLayerForChecksum(ipPacket); err != nil {
		return 0, err
	}

	// Serialize.  Note:  we only serialize the TCP layer, because the
	// socket we get with net.ListenPacket wraps our data in IPv4 packets
	// already.  We do still need the IP layer to compute checksums
	// correctly, though.
	ipHeaderBuf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if err = ipPacket.SerializeTo(ipHeaderBuf, opts); err != nil {
		return 0, err
	}

	if ipHeader, err = ipv4.ParseHeader(ipHeaderBuf.Bytes()); err != nil {
		return 0, err
	}

	ethernetLayer := buildEthernetPacket(c.Ethernet)
	tcpPayloadBuf := gopacket.NewSerializeBuffer()
	pyl := gopacket.Payload(c.Payload)

	if err = gopacket.SerializeLayers(tcpPayloadBuf, opts, ethernetLayer, tcpPacket, pyl); err != nil {
		return 0, err
	}

	// XXX send packet
	if packetConn, err = net.ListenPacket("ip4:tcp", "0.0.0.0"); err != nil {
		return 0, err
	}

	if rawConn, err = ipv4.NewRawConn(packetConn); err != nil {
		return 0, err
	}

	if err = rawConn.WriteTo(ipHeader, tcpPayloadBuf.Bytes(), nil); err != nil {
		return 0, err
	}
	return len(c.Payload), nil
}

type IPPacketConfig struct {
	SrcIP string `json:"src_ip"`
	DstIP string `json:"dst_ip"`
}

// buildIpPacket generates a layers.IPv4 and returns it with source IP address and destination IP address
func buildIpPacket(c IPPacketConfig) *layers.IPv4 {
	return &layers.IPv4{
		SrcIP:    net.ParseIP(c.SrcIP).To4(),
		DstIP:    net.ParseIP(c.DstIP).To4(),
		Version:  4,
		Protocol: layers.IPProtocolTCP,
	}
}

type TCPFlagsConfig struct {
	SYN bool
	ACK bool
	FIN bool
	RST bool
	PSH bool
	URG bool
	ECE bool
	CWR bool
	NS  bool
}

type TCPPacketConfig struct {
	SrcPort int `json:"src_port"`
	DstPort int `json:"dst_port"`
	Seq     uint32
	Ack     uint32
	Window  uint16
	Urgent  uint16
	Flags   TCPFlagsConfig
}

// buildTcpPacket generates a layers.TCP and returns it with source port and destination port
func buildTcpPacket(c TCPPacketConfig) *layers.TCP {
	return &layers.TCP{
		SrcPort: layers.TCPPort(c.SrcPort),
		DstPort: layers.TCPPort(c.DstPort),
		Window:  c.Window,
		Urgent:  c.Urgent,
		Seq:     c.Seq,
		Ack:     c.Ack,
		SYN:     c.Flags.SYN,
		ACK:     c.Flags.ACK,
		FIN:     c.Flags.FIN,
		RST:     c.Flags.RST,
		PSH:     c.Flags.PSH,
		URG:     c.Flags.URG,
		ECE:     c.Flags.ECE,
		CWR:     c.Flags.CWR,
		NS:      c.Flags.NS,
	}
}

type EthernetPacketConfig struct {
	SrcMAC []byte `json:"src_mac"`
	DstMAC []byte `json:"dst_mac"`
}

// buildEthernetPacket generates an layers.Ethernet and returns it with source MAC address and destination MAC address
func buildEthernetPacket(c EthernetPacketConfig) *layers.Ethernet {
	return &layers.Ethernet{
		SrcMAC: net.HardwareAddr{c.SrcMAC[0], c.SrcMAC[1], c.SrcMAC[2], c.SrcMAC[3], c.SrcMAC[4], c.SrcMAC[5]},
		DstMAC: net.HardwareAddr{c.DstMAC[0], c.DstMAC[1], c.DstMAC[2], c.DstMAC[3], c.DstMAC[4], c.DstMAC[5]},
	}
}
