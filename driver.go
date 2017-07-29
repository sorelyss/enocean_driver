package main

import (
  "fmt"
  "github.com/tarm/serial"
  "time"
  "strings"
  "strconv"
)

var crcTable = []byte{
    0x00, 0x07, 0x0e, 0x09, 0x1c, 0x1b, 0x12, 0x15,
    0x38, 0x3f, 0x36, 0x31, 0x24, 0x23, 0x2a, 0x2d,
    0x70, 0x77, 0x7e, 0x79, 0x6c, 0x6b, 0x62, 0x65,
    0x48, 0x4f, 0x46, 0x41, 0x54, 0x53, 0x5a, 0x5d,
    0xe0, 0xe7, 0xee, 0xe9, 0xfc, 0xfb, 0xf2, 0xf5,
    0xd8, 0xdf, 0xd6, 0xd1, 0xc4, 0xc3, 0xca, 0xcd,
    0x90, 0x97, 0x9e, 0x99, 0x8c, 0x8b, 0x82, 0x85,
    0xa8, 0xaf, 0xa6, 0xa1, 0xb4, 0xb3, 0xba, 0xbd,
    0xc7, 0xc0, 0xc9, 0xce, 0xdb, 0xdc, 0xd5, 0xd2,
    0xff, 0xf8, 0xf1, 0xf6, 0xe3, 0xe4, 0xed, 0xea,
    0xb7, 0xb0, 0xb9, 0xbe, 0xab, 0xac, 0xa5, 0xa2,
    0x8f, 0x88, 0x81, 0x86, 0x93, 0x94, 0x9d, 0x9a,
    0x27, 0x20, 0x29, 0x2e, 0x3b, 0x3c, 0x35, 0x32,
    0x1f, 0x18, 0x11, 0x16, 0x03, 0x04, 0x0d, 0x0a,
    0x57, 0x50, 0x59, 0x5e, 0x4b, 0x4c, 0x45, 0x42,
    0x6f, 0x68, 0x61, 0x66, 0x73, 0x74, 0x7d, 0x7a,
    0x89, 0x8e, 0x87, 0x80, 0x95, 0x92, 0x9b, 0x9c,
    0xb1, 0xb6, 0xbf, 0xb8, 0xad, 0xaa, 0xa3, 0xa4,
    0xf9, 0xfe, 0xf7, 0xf0, 0xe5, 0xe2, 0xeb, 0xec,
    0xc1, 0xc6, 0xcf, 0xc8, 0xdd, 0xda, 0xd3, 0xd4,
    0x69, 0x6e, 0x67, 0x60, 0x75, 0x72, 0x7b, 0x7c,
    0x51, 0x56, 0x5f, 0x58, 0x4d, 0x4a, 0x43, 0x44,
    0x19, 0x1e, 0x17, 0x10, 0x05, 0x02, 0x0b, 0x0c,
    0x21, 0x26, 0x2f, 0x28, 0x3d, 0x3a, 0x33, 0x34,
    0x4e, 0x49, 0x40, 0x47, 0x52, 0x55, 0x5c, 0x5b,
    0x76, 0x71, 0x78, 0x7f, 0x6A, 0x6d, 0x64, 0x63,
    0x3e, 0x39, 0x30, 0x37, 0x22, 0x25, 0x2c, 0x2b,
    0x06, 0x01, 0x08, 0x0f, 0x1a, 0x1d, 0x14, 0x13,
    0xae, 0xa9, 0xa0, 0xa7, 0xb2, 0xb5, 0xbc, 0xbb,
    0x96, 0x91, 0x98, 0x9f, 0x8a, 0x8D, 0x84, 0x83,
    0xde, 0xd9, 0xd0, 0xd7, 0xc2, 0xc5, 0xcc, 0xcb,
    0xe6, 0xe1, 0xe8, 0xef, 0xfa, 0xfd, 0xf4, 0xf3}

type USB300 struct {
  dongle_id []byte
  ser *serial.Port
  states []bool
  last_commits []int64
  switch_names []string
  switch_modes []int64
}

type Info struct {
  State bool
  Last_commit int64
  Switch_mode int64
  Switch_name string
}

func NewUSB300(dongle_id string, port string) (*USB300, error) {
  c := &serial.Config{Name: port, Baud: 57600}
  ser, err := serial.OpenPort(c)
  if err != nil {
      return nil, fmt.Errorf("Failed connecting the USB300: %v", err)
  }
  hex_array := strings.Split(dongle_id,":")

  base_id := []byte{}
  for i := 0; i < len(hex_array); i++ {
    hex_Int, _ := strconv.ParseInt(hex_array[i],16,16)  
    base_id = append(base_id, byte(hex_Int))
  }

  usb300 := USB300 {
    ser: ser,
    dongle_id: []byte(base_id),
    states: []bool{}, 
    last_commits: []int64{},
    switch_names: []string{},
    switch_modes: []int64{},
  }
  return &usb300,err
}

func (v *USB300) ActuateSwitch(switch_index int, state bool) {

  if state == false {
    turn_off(v, switch_index)
  } else {
    turn_on(v, switch_index)
  }
  fmt.Println()
  v.states[switch_index] = state
  v.last_commits[switch_index] = time.Now().UnixNano()
}

func (v *USB300) SetSwitches(switch_modes []string, switch_names []string) {
  for i := 0; i < len(switch_modes); i++ {
    temp, _ := time.ParseDuration( switch_modes[i] )
    v.switch_modes = append(v.switch_modes,temp.Nanoseconds())
  }
  v.switch_names = switch_names  
  v.last_commits = make([]int64, len(switch_modes))
  v.states = make([]bool, len(switch_modes))
}


func (v *USB300) GetStatus() []Info {
  var status []Info
  for i := 0; i < len(v.switch_names); i++ {
    if(time.Now().UnixNano() - v.last_commits[i] > v.switch_modes[i]){
      v.states[i] = false
    }
    temp := Info{ State: v.states[i], Last_commit: v.last_commits[i], Switch_name: v.switch_names[i], Switch_mode: v.switch_modes[i],}
    status = append(status, temp)
    fmt.Println("Switch "+temp.Switch_name+": {state:", temp.State,", mode:",temp.Switch_mode,"}")
    }
    fmt.Println()  
  return status
}

func send4BTelegram(v *USB300, tx_index int, data []byte) {
  // Send the data array through the serial port 
  // using the input id index
  //    index: (Integer) index of the control node
  var crc uint8
  data = append(data, getTx_ID(byte(tx_index), 3, v.dongle_id)...)
  for index := range data {
    crc = crcTable[(crc^data[index]) & 0xff ]
  }
  telegram := append([]byte("\x55\x00\x0A\x00\x01\x80"), data...)
  telegram = append(telegram, crc)
  v.ser.Write(telegram)
  //ser.Close()
}

func getTx_ID(Iface_index byte, index int, base_id []byte) []byte {
  // Get the base id of the interface and add status byte
  //    tx_id: (Integer) index of the interface
  sum := base_id
  before_overfw := byte(255 - base_id[index])
  if Iface_index >= before_overfw {
    sum[index] = 255
    return getTx_ID(Iface_index - before_overfw, index-1, sum)
  }else{  
    sum[index] = sum[index] + Iface_index
    return append(sum, 0x80)
  } 
}

func turn_on(v *USB300, index int) {
  // Send the commands to turn on the lights paired with the index
  //    index: (Integer) index of the control node
  send4BTelegram(v, index, []byte{0xA5, 0x1C, 0x10, 0x1B, 0x84} ) // 84 Test
  send4BTelegram(v, index, []byte{0xA5, 0x05, 0x00, 0x0B, 0x88} ) // 88 Motion detected
  time.Sleep(1000 * time.Millisecond)
  send4BTelegram(v, index, []byte{0xA5, 0x1C, 0x10, 0x1B, 0x80} ) // 80 Pair
  fmt.Println("Lights - Iface ",index," - turned on - Wait 4 secs")
}

func turn_off(v *USB300, index int) {
  // Send the commands to turn on the lights paired with the index
  //    index: (Integer) index of the control node
  send4BTelegram(v, index, []byte{0xA5, 0x1C, 0x10, 0x1B, 0x84} ) // 84 Test
  time.Sleep(1000 * time.Millisecond)
  send4BTelegram(v, index, []byte{0xA5, 0x1C, 0x10, 0x1B, 0x80} ) // 80 Pair
  fmt.Println("Lights - Iface ",index," - turned off - Wait 4 secs")
}
