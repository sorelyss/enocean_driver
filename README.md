# Enocean Driver for BW2
This repository contains the BOSSWAVE driver for the wattstopper switches using the Enocean USB300 dongle. 
The driver supports the creations of multiple interfaces, that means multiple switches controlled by the same dongle.
The repository also includes a python file that helps setting up the params.yml file and the pairing of the switches with the USB300 dongle.
Finally there is a library for the integration with Bodge in order to make easier the control management.

## Requirements
- <a href="https://golang.org/dl/">Go</a>
- <a href="https://www.python.org/downloads/">Python 2.7</a>
- <a href="https://github.com/immesys/bw2">BW2 (curl get.bw2.io/agent | sh)</a>
- <a href="https://github.com/gtfierro/bodge">Python 2.7</a>
## Setup
To setup your machine the \~/.bashrc file should look like:
```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=~/go
export PATH=$PATH:$GOPATH/bin
export BW2_DEFAULT_ENTITY=~/PATH_TO_THE_ENTITY/NAME_OF_THE_ENTITY.ent
```
Pairing the Dongle with the lights:
```bash
sudo pip2 install pyyaml pyserial
python2 pair_enocean.py
```
How the params.yml file should look like:
```yml
port: /dev/ttyUSB0
svc_base_uri: scratch.ns/light_test
name: kitchen
USB300_id: ff:c7:04:80
poll_interval: 10s
switch_names: [main_lights, second_lights, fan, third_lights]
switch_modes: [15m, 15m, 15m, 15m]
```
In order to add the enocean library to bodge, run:
```bash
bodge publish xbos_enocean.lua bodge/contrib/xbos_enocean.lua
```
## Example
The next example is going to use two terminals, the first will run the driver and the second will handle the actuation
signals by creating an object of the class xbos_enocean for each interface that wants to be controlled.
Each object has the methods: 
- state(): _retrieve the state of the interface_
- state(0/1): _set the state of the interface_
- last_commit(): _retrieve the last commit registered of the interface_
- switch_name(): _retrieve the name of the interface_
- switch_mode(): _retrieve the mode of the interface_

Terminal 1:
```bash
sudo service bw2 start
bw2 status # ready when current_block==seen_block && peer_count>0
cd enocean; go build; 
./enocean
```

Terminal 2:
```lua
enocean_class = bw.uriRequire("bodge/contrib/xbos_enocean.lua")
sw1 = enocean_class("scratch.ns/light_test/kitchen/s.enocean/0/i.xbos.light")
sw2 = enocean_class("scratch.ns/light_test/kitchen/s.enocean/1/i.xbos.light")
sw4 = enocean_class("scratch.ns/light_test/kitchen/s.enocean/3/i.xbos.light")
sw4:state(1); sw2:state(1); sw1:state(1)
sw1:state(0); sw2:state(0); sw4:state(0)
```

