package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"time"

	"github.com/fatih/color"
)

const symbols = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_"

func FindTagValue(b []byte, tag string) string {
	re := regexp.MustCompile(`(?s)[:<]` + tag + `>([^<]+)`)
	m := re.FindSubmatch(b)
	if len(m) != 2 {
		return ""
	}
	return string(m[1])
}

func RandString(size, base byte) string {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	if base == 0 {
		return string(b)
	}
	for i := byte(0); i < size; i++ {
		b[i] = symbols[b[i]%base]
	}
	return string(b)
}

// UUID - generate something like 44302cbf-0d18-4feb-79b3-33b575263da3
func UUID() string {
	s := RandString(32, 16)
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

// ListInterfaces lists all available network interfaces
func ListInterfaces() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %v", err)
	}
	return interfaces, nil
}

// DiscoveryStreamingURLs discovers ONVIF streaming URLs on the specified interface
func DiscoveryStreamingURLs(interfaceName string) ([]string, error) {
	// Resolve the multicast address
	addr := &net.UDPAddr{
		IP:   net.IP{239, 255, 255, 250},
		Port: 3702,
	}

	// Get the network interface
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("unable to find interface %s: %v", interfaceName, err)
	}

	// Bind the UDP socket to the interface
	conn, err := net.ListenMulticastUDP("udp4", iface, addr)
	if err != nil {
		return nil, fmt.Errorf("unable to bind to multicast address on %s: %v", interfaceName, err)
	}
	defer conn.Close()

	// Set a read deadline
	if err = conn.SetReadDeadline(time.Now().Add(8 * time.Second)); err != nil {
		return nil, err
	}

	// WS-Discovery message
	msg := `<?xml version="1.0" ?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Header xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing">
		<a:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe</a:Action>
		<a:MessageID>urn:uuid:` + UUID() + `</a:MessageID>
		<a:To>urn:schemas-xmlsoap-org:ws:2005:04/discovery</a:To>
	</s:Header>
	<s:Body>
		<d:Probe xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery">
			<d:Types />
			<d:Scopes />
		</d:Probe>
	</s:Body>
</s:Envelope>`

	// Send the WS-Discovery message
	if _, err = conn.WriteTo([]byte(msg), addr); err != nil {
		return nil, fmt.Errorf("error sending discovery message: %v", err)
	}

	var urls []string
	buffer := make([]byte, 8192)

	for {
		// Read responses
		n, respAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return nil, fmt.Errorf("error reading from UDP: %v", err)
		}

		response := buffer[:n]

		// Ignore non-ONVIF responses
		if !strings.Contains(string(response), "onvif") {
			continue
		}

		// Extract the XAddrs tag value
		url := FindTagValue(response, "XAddrs")
		if url == "" {
			continue
		}

		// Fix URLs with "http://0.0.0.0"
		if strings.HasPrefix(url, "http://0.0.0.0") {
			url = "http://" + respAddr.IP.String() + url[14:]
		}

		urls = append(urls, url)
	}

	return urls, nil
}

func main() {
	// Step 1: List all network interfaces
	interfaces, err := ListInterfaces()
	if err != nil {
		fmt.Printf("Error listing interfaces: %v\n", err)
		return
	}

	if len(interfaces) == 0 {
		fmt.Println("No network interfaces found.")
		return
	}

	fmt.Println(color.CyanString("Available network interfaces:"))
	for i, iface := range interfaces {
		fmt.Printf("[%d] %v (Flags: %s)\n", i, color.GreenString(iface.Name), iface.Flags)

	}

	// Step 2: Ask the user to select an interface
	fmt.Print(color.CyanString("Select an interface (enter the number): "))
	reader := bufio.NewReader(os.Stdin)
	selection, _ := reader.ReadString('\n')
	selection = strings.TrimSpace(selection)

	index, err := strconv.Atoi(selection)
	if err != nil || index < 0 || index >= len(interfaces) {
		fmt.Println("Invalid selection.")
		return
	}

	selectedInterface := interfaces[index].Name
	fmt.Printf("Selected interface: %s\n", selectedInterface)

	// Step 3: Perform ONVIF discovery on the selected interface
	urls, err := DiscoveryStreamingURLs(selectedInterface)
	if err != nil {
		fmt.Printf("Error discovering streaming URLs: %v\n", err)
		return
	}

	// Step 4: Print discovered URLs
	if len(urls) == 0 {
		fmt.Println("No ONVIF devices found.")
	} else {
		fmt.Println("Discovered ONVIF streaming URLs:")
		for _, url := range urls {
			color.Yellow(url)
			// fmt.Println(url)
		}
	}
}
