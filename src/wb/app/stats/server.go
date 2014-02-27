package stats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
	. "wb/app/config"
)

const (
	MAX_UDP_PACKET_SIZE = 2048
	MAX_PACKET_HISTORY  = 1000
)

type Packet struct {
	Time time.Time `json:"time"`
	Type string    `json:"type"`
	Data string    `json:"data"`
	Address string `json:"address"`
}

var Packets []Packet

func StartUDPServer() {
	go func() {
		address, err := net.ResolveUDPAddr("udp", UDPAddress)
		listener, err := net.ListenUDP("udp", address)
		if err != nil {
			ERROR.Println("Cannot listen using UDP - %s", err)
		}
		defer listener.Close()

		message := make([]byte, MAX_UDP_PACKET_SIZE)
		for {
			if n, _, err := listener.ReadFromUDP(message); err == nil && n > 0 {
				if len(Packets) > MAX_PACKET_HISTORY {
					Packets = Packets[:len(Packets)/2]
				}
				data := message[:n]
				var packet Packet
				if err := json.Unmarshal(data, &packet); err == nil {
					Packets = append([]Packet{packet}, Packets...)
				}
			}
		}
	}()
}

func SendMessage(cat string, msg string) {
	packet := Packet{
		Time: time.Now(),
		Type: cat,
		Data: msg,
		Address:  Address,
	}
	if EnableUDP {
		go func() {
			Log(cat, msg)
			if connection, err := net.Dial("udp", UDPAddress); err == nil {
				defer connection.Close()
				js, err := json.Marshal(packet)
				if err == nil {
					fmt.Fprintf(connection, string(js))
				}
			}
		}()
	} else {
		if js, err := json.Marshal(packet); err == nil {
			b := bytes.NewBufferString(string(js))
			addr := MasterAddress + "/log_message"
			resp, _ := http.Post(addr, "text/json", b)
			defer resp.Body.Close()
		}
	}
}
