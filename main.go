package main

import (
	"fmt"
	"time"
	"os"
	"strconv"

	"github.com/immesys/spawnpoint/spawnable"
	bw2 "gopkg.in/immesys/bw2bind.v5"
)

const (
	PONUM = "2.1.1.2"
)

type XBOSEnoceanReading struct {
	Time  		int64	`msgpack:"time"`
	State 		bool	`msgpack:"state"`
	Last_commit int64	`msgpack:"last_commit"`
	Switch_name string	`msgpack:"switch_name"`
	Switch_mode int64	`msgpack:"switch_mode"`
}

func (tpl *XBOSEnoceanReading) ToMsgPackPO() bw2.PayloadObject {
	po, err := bw2.CreateMsgPackPayloadObject(bw2.FromDotForm(PONUM), tpl)
	if err != nil {
		panic(err)
	} else {
		return po
	}
}

func main() {
	bwClient := bw2.ConnectOrExit("")

	params := spawnable.GetParamsOrExit()
	bwClient.OverrideAutoChainTo(true)
	bwClient.SetEntityFromEnvironOrExit()

	baseURI := params.MustString("svc_base_uri")
	poll_interval := params.MustString("poll_interval")
	pollInt, err := time.ParseDuration(poll_interval)
	if err != nil {
		panic(err)
	}
	deviceName := params.MustString("name")
	if deviceName == "" {
		os.Exit(1)
	}

	service := bwClient.RegisterService(baseURI+"/"+deviceName, "s.enocean")

	port_name := params.MustString("port")
	USB300_id := params.MustString("USB300_id")
	v, err := NewUSB300(USB300_id, port_name)
	if err != nil {
		panic(err)
	}
	switch_modes := params.MustStringSlice("switch_modes")
	switch_names := params.MustStringSlice("switch_names")
	v.SetSwitches(switch_modes, switch_names)
	fmt.Printf("Initialized driver for USB300 dongle\n")

	xbosIface := make([]*bw2.Interface, len(switch_names))
	for i := 0; i < len(switch_names); i++ {
		xbosIface[i] = service.RegisterInterface(strconv.Itoa(i), "i.xbos.light")
	}

	for j := 0; j < len(switch_names); j++ {
		idx := j
		xbosIface[j].SubscribeSlot("state", func(msg *bw2.SimpleMessage) {
			po := msg.GetOnePODF(PONUM)
			if po == nil {
				fmt.Println("Received actuation command without valid PO, dropping")
				return
			} else if len(po.GetContents()) < 1 {
				fmt.Println("Received actuation command with invalid PO, dropping")
				return
			}
			
			msgpo, err := bw2.LoadMsgPackPayloadObject(po.GetPONum(), po.GetContents())
			if err != nil {
				fmt.Println(err)
				return
			}
			var data map[string]interface{}

			err = msgpo.ValueInto(&data)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Actuation signal:",data, "@ Iface_index: ", idx)
			state := int(data["state"].(float64))
			v.ActuateSwitch(idx, state != 0)
		})
	}

	// Publish status information
	go func() {
		for {
			switchStatuses := v.GetStatus()

			timestamp := time.Now().UnixNano()
			for i := 0; i < len(switchStatuses); i++ {
				msg := &XBOSEnoceanReading{
					Time:  timestamp,
					State: switchStatuses[i].State,
					Last_commit: switchStatuses[i].Last_commit,
					Switch_name: switchStatuses[i].Switch_name,
					Switch_mode: switchStatuses[i].Switch_mode,
				}
				po := msg.ToMsgPackPO()
				xbosIface[i].PublishSignal("info", po)
			}

			time.Sleep(pollInt)
		}

	}()
	
	for {
		time.Sleep(10 * time.Second)
	}

}

