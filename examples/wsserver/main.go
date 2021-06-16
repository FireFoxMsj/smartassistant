package main

import (
	"github.com/gorilla/websocket"
	"gitlab.yctc.tech/root/smartassistent.git/core"
	"log"
	"net/url"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8088", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// set_bright
	// cs := core.CallService{
	// 	Domain:      "yeelight",
	// 	ID:          1,
	// 	Service:     "set_bright",
	// 	ServiceData: core.M{"brightness_pct": 80},
	// }

	cs := core.CallService{
		Domain:  "plugin",
		ID:      1,
		Service: "install",
		ServiceData: core.M{
			"download_url": "http://tysq2.yctc.tech/api/file/originals/id/2009037/fn/plugin1.zip",
			"name":         "plugin3",
		},
	}

	if err := c.WriteJSON(&cs); err != nil {
		panic(err)
	}

	m1 := make(map[string]interface{})

	m1["asdf"] = "sdfs"
	m1["23"] = 34
	for {
		if _, res, err := c.ReadMessage(); err != nil {
			break
		} else {
			log.Println(string(res))
		}
	}

}
