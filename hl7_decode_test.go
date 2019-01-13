// Unit tests for HL7 v2 decoding

package main

import (
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func TestHL7DecodeFile(t *testing.T) {
	handle, err := pcap.OpenOffline("testdata/HL7-ADT-UDI-PRT.pcap")
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		app := packet.ApplicationLayer()
		if app == nil {
			continue // Ignore packets without an application layer
		}

		_, _, err := hl7Decode(&app)
		if err != nil {
			panic(err)
		}
	}
}

func appLayerFromString(s string) *gopacket.ApplicationLayer {
	bytes := []byte(s)
	appLayer := gopacket.ApplicationLayer(gopacket.Payload(bytes))
	return &appLayer
}

func TestHL7DecodeTooShort(t *testing.T) {
	appLayer := appLayerFromString("MSH")
	ident, _, err := hl7Decode(appLayer)
	if ident != "" {
		t.Errorf("Got identifier when none was expected")
	}
	if err == nil {
		t.Errorf("Expected an error from too-short HL7 message")
	}
}

func testHL7DecodeEmpty(s string, t *testing.T) {
	appLayer := appLayerFromString(s)
	ident, _, err := hl7Decode(appLayer)
	if ident != "" {
		t.Errorf("Got identifier when none was expected")
	}
	if err != nil {
		panic(err)
	}
}

func TestHL7DecodeEmpty1(t *testing.T) { testHL7DecodeEmpty("MSH|^~\\&", t) }
func TestHL7DecodeEmpty2(t *testing.T) { testHL7DecodeEmpty("MSH|^~\\&|", t) }

func identFromString(s string) string {
	appLayer := appLayerFromString(s)
	ident, _, err := hl7Decode(appLayer)
	if err != nil {
		panic(err)
	}
	return ident
}

// Well-formed message header segment to be prepended to messages for testing
const okHl7Header = ("" +
	// Header and delimiter
	"MSH|^~\\&|" +

	// Envelope information
	"Sender|Sender Facility|" +
	"Receiver|Receiver Facility|" +

	// Timestamp (YYYYMMDDHHMM) + Security (blank)
	"201801131030||" +

	// Message type: ORU = observations & results
	"ORU^R01|" +

	// Control ID
	"CNTRL-12345|" +

	// Processing ID
	"P|" +

	// Version ID + segment delimiter (carriage return)
	"2.4\r")

func TestHL7IdentFromPRT10(t *testing.T) {
	str := (okHl7Header +
		"PRT|A|B|C|D|E|F|G|H|I|Grospira Peach B+\r")
	parsed := identFromString(str)
	if parsed != "Grospira Peach B+" {
		t.Errorf("Failed to parse identifier from string; got '%s'", parsed)
	}
}

func BenchmarkHL7IdentFromPRT10(b *testing.B) {
	str := (okHl7Header +
		"PRT|A|B|C|D|E|F|G|H|I|Grospira Peach B+\r")
	for i := 0; i < b.N; i++ {
		identFromString(str)
	}
}