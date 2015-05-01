package stats

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	. "wb/app/config"
)

const (
	MAX_PACKET_HISTORY = 1000
)

type Packet struct {
	Time    time.Time `json:"time"`
	Type    string    `json:"type"`
	Data    string    `json:"data"`
	Address string    `json:"address"`
}

var Packets []Packet

func SendMessage(cat string, msg string) {
	packet := Packet{
		Time:    time.Now(),
		Type:    cat,
		Data:    msg,
		Address: Address,
	}
	if js, err := json.Marshal(packet); err == nil {
		b := bytes.NewBufferString(string(js))
		addr := MasterAddress + "/log_message"
		resp, _ := http.Post(addr, "text/json", b)
		defer resp.Body.Close()
	}
}
