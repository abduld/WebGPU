package controllers

import (
	"wb/app/stats"

	"code.google.com/p/go.net/websocket"
	"github.com/garyburd/redigo/redis"
	"github.com/robfig/revel"
)

type Dashboard struct {
	*revel.Controller
}

func (c Dashboard) DashboardSocket(ws *websocket.Conn) revel.Result {

	type Packet struct {
		Channel string `json:"channel"`
		Data    string `json:"data"`
	}

	onPMessageRecieve := func(m redis.PMessage, ch string, psc redis.PubSubConn) error {
		pkt := Packet{
			Channel: m.Channel,
			Data:    string(m.Data),
		}
		if err := websocket.JSON.Send(ws, &pkt); err != nil {
			revel.TRACE.Println("Got message in socket...")
			return err
		}
		return nil
	}
	onMessageRecieve := func(m redis.Message, ch string, psc redis.PubSubConn) error {
		pkt := Packet{
			Channel: m.Channel,
			Data:    string(m.Data),
		}
		if err := websocket.JSON.Send(ws, &pkt); err != nil {
			revel.TRACE.Println("Got message in socket...")
			return err
		}
		return nil
	}

	onSubscription := func(sub redis.Subscription, ch string, psc redis.PubSubConn) error {
		return nil
	}

	onError := func(err error, ch string, psc redis.PubSubConn) error {
		return err
	}
	stats.RedisSubscribe("*", onMessageRecieve, onPMessageRecieve, onSubscription, onError)

	return nil
}

func (c Dashboard) ViewDashboard() revel.Result {
	return c.Render()
}
