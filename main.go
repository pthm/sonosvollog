package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/ianr0bkny/go-sonos"
	"github.com/ianr0bkny/go-sonos/ssdp"
)

type Punish struct {
	Enabled bool
	Threshold int
	Ideal int
}

func logVolume(dev ssdp.Device, punish Punish) {
	dp := sonos.Connect(dev, nil, sonos.SVC_DEVICE_PROPERTIES)
	rc := sonos.Connect(dev, nil, sonos.SVC_RENDERING_CONTROL)
	name, _, _ := dp.DeviceProperties.GetZoneAttributes()

	fmt.Printf("Logging volume for %v\n", name)

	ticker := time.NewTicker(10 * time.Second)
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	file, err := os.OpenFile("./log.csv", os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(file)
	go func(s *sonos.Sonos) {
		for {
			select {
			case <-done:
				log.Printf("Closing")
				return
			case t := <-ticker.C:
				vol, err := s.RenderingControl.GetVolume(0,"Master")
				if err != nil {
					panic(err)
				}
				fmt.Printf("Vol @ %v: %v\n", t.Format("2006-01-02 15:04:05"), vol)
				err = w.Write([]string{t.Format("2006-01-02 15:04:05"), fmt.Sprintf("%v",vol)})
				if err != nil {
					panic(err)
				}
				if punish.Enabled && vol > uint16(punish.Threshold) {
					fmt.Printf("PUNISH: Volume exceeded threshold of %v lowering it to %v\n", punish.Threshold, punish.Ideal)
					s.RenderingControl.SetVolume(0, "Master", uint16(punish.Ideal))
				}
			}
		}
	}(rc)

	<-done
}

type LoggerDevice struct {
	Name string
	Dev ssdp.Device
}

func readInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%v: ", prompt)
	text, _ := reader.ReadString('\n')

	return strings.TrimRight(text, "\n")
}

func main() {
	log.SetOutput(ioutil.Discard)

	punish := flag.Bool("punish", false, "If volume exceeds threshold lower it")
	threshold := flag.Int("threshold", 30, "Threshold to use for punishment (default: 30)")
	ideal := flag.Int("ideal", 20, "Volume to set as punishment (default: 20)")
	flag.Parse()

	spew.Dump(punish, threshold, ideal)
	punishConfig := Punish{
		Enabled:   *punish,
		Threshold: *threshold,
		Ideal:     *ideal,
	}

	mgr := ssdp.MakeManager()

	fmt.Println("Choose a network interface")
	ints, err := net.Interfaces()
	if err != nil {
		panic(err)
	}
	for idx, int := range ints {
		fmt.Printf("\t%v. %v\n", idx+1, int.Name)
	}
	choiceStr := readInput("Choose (Enter number)")
	choiceIdx, err := strconv.Atoi(choiceStr)
	if err != nil {
		panic(err)
	}
	choiceIdx--
	chosenInt := ints[choiceIdx]

	fmt.Println("Searching for Sonos devices...")
	if err := mgr.Discover(chosenInt.Name, "11209", false); err != nil {
		panic(err)
	}

	// SericeQueryTerms
	// A map of service keys to minimum required version
	qry := ssdp.ServiceQueryTerms{
		ssdp.ServiceKey("schemas-upnp-org-MusicServices"): -1,
	}
	var devs []LoggerDevice
	// Look for the service keys in qry in the database of discovered devices
	result := mgr.QueryServices(qry)
	if dev_list, has := result["schemas-upnp-org-MusicServices"]; has {
		for _, dev := range dev_list {
			dp := sonos.Connect(dev, nil, sonos.SVC_DEVICE_PROPERTIES)
			name, _, _ := dp.DeviceProperties.GetZoneAttributes()
			devs = append(devs, LoggerDevice{
				Name: name,
				Dev: dev,
			})
		}
	}

	fmt.Printf("Found %v Sonos devices:\n", len(devs))
	for idx, dev := range devs {
		fmt.Printf("\t%v. %v\n", idx+1, dev.Name)
	}
	choiceStr = readInput("Choose (Enter number)")
	choiceIdx, err = strconv.Atoi(choiceStr)
	if err != nil {
		panic(err)
	}
	choiceIdx--

	if punishConfig.Enabled {
		fmt.Println("Punishment enabled!")
		fmt.Printf("If volume exceeds %v it will be lowered to %v\n", punishConfig.Threshold, punishConfig.Ideal)
	}
	logVolume(devs[choiceIdx].Dev, punishConfig)

	mgr.Close()
}