package main

import (
	"errors"
	"fmt"
	"gitlab.yctc.tech/root/smartassistent.git/core"
	"math/rand"
	"time"
)

func main() {
	sr := core.NewSmartAssistant()
	sr.Services.Register("domain", "turn_on", handle, schema)
	sr.Bus.Listen("yeelight.turn_on", TurnOnCB)
	eventData := map[string]interface{}{"test1": "abcd", "test2": "efgt"}
	err := sr.Services.Call("domain", "turn_on", eventData)
	errHandler(err)

	eventData = map[string]interface{}{"domain": "domain", "service": "turn_on"}
	err = sr.Bus.Fire("yeelight.turn_on", eventData)
	errHandler(err)

	//p, err := plugin.Open("/Users/lanrion/Yctc/code/gitlab.yctc.tech/root/smartassistent.git/plugins/firstplugin/plugins.so")
	//errHandler(err)
	//
	//sy, err := p.Lookup("Register")
	//errHandler(err)
	//sy.(func(sa *core.SmartAssistant))(sr)
	eventData1 := map[string]interface{}{"test1": "abcd123123123", "test2": "efgtasdfsfd"}
	err = sr.Services.Call("pluginTest", "turn_on", eventData1)
	errHandler(err)
}

func TurnOnCB(event core.Event) error {
	fmt.Printf("type: %s, data: %v", event.EventType, event.Data)
	return errors.New("test err")
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func errHandler(err error) {
	if err != nil {
		panic(err)
	}
}

func handle(args core.M) error {
	fmt.Println(args)
	return nil
}

func schema(args core.M) core.M {
	str := make([]string, 0)

	for _, arg := range args {
		fmt.Println(arg)
		str = append(str, arg.(string))
	}

	var sd = make(core.M)
	sd["res"] = "123"
	return sd
}
