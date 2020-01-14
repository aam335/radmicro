package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"

	radius "github.com/aam335/go-radius"
)

// ParceUDPAddr host || host:port to net.UDPAddr
func ParceUDPAddr(s string) (*net.UDPAddr, error) {
	dst := &net.UDPAddr{}
	if dst.IP = net.ParseIP(s); dst.IP != nil {
		return dst, nil
	}
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return nil, err
	}
	if dstPort, err := strconv.ParseUint(port, 10, 16); err == nil {
		dst.Port = int(dstPort)
	}
	if dst.IP = net.ParseIP(host); dst.IP == nil {
		return nil, errors.New("invalid address format")
	}
	return dst, nil
}

func usage(s string, err error) {
	flag.Usage()
	// var buff bytes.Buffer
	// e := toml.NewEncoder(&buff)
	// pkt := RadFields{Send: RadField{Code: radCode{radius.CodeAccessRequest}, Attrs: map[string]string{"attr1": "val1", "attr2": "val2"}}}
	// if err = e.Encode(&pkt); err == nil {
	// 	fmt.Print(string(buff.Bytes()))
	// 	os.Exit(0)
	// }
	log.Fatalf("Error: %v %v", s, err)
}

func main() {
	// log.Println(os.Args)
	var serverIPPort = flag.String("serv", "127.0.0.1:1812", "Radius adress:port")
	var localIPPort = flag.String("me", "", "Source adress:port")
	var timeoutStr = flag.String("t", "1s", "timeout")
	var retries = flag.Int("r", 1, "retries")
	var secret = flag.String("secret", "", "Radius secret")

	var verboseReply = flag.Bool("v", false, "Verbose radius reply")
	var verboseJSON = flag.Bool("vj", false, "Verbose radius reply vin json")
	var qt = flag.Bool("qt", false, "Request execution time")

	var packetFile = flag.String("p", "", "packet file: packet.[json|toml|yaml]")

	flag.Parse()

	if *secret == "" {
		usage("secret", errors.New("must be set"))
	}

	var rf RadFields
	if err := cleanenv.ReadConfig(*packetFile, &rf); err != nil {
		usage(*packetFile, err)
	}

	tmo, err := time.ParseDuration(*timeoutStr)
	if err != nil {
		usage("timeout ", err)
	}

	client := radius.Client{Timeout: tmo, Retries: *retries}
	if *localIPPort != "" {
		src, err := ParceUDPAddr(*localIPPort)
		if err != nil {
			usage("local address ", err)
		}
		client.LocalAddr = src
	}

	packet := radius.New(rf.Send.Code.Code, []byte(*secret))
	// packet.Add("Calling-Station-Id", "NAS-Fake")
	//	rx := regexp.MustCompile("\\s*:\\s*")
	for name, value := range rf.Send.Attrs {
		if err := packet.Add(name, value); err != nil {
			log.Fatalf("'%v':'%v' error:%v", name, value, err)
		}
	}

	dst, err := ParceUDPAddr(*serverIPPort)
	if err != nil {
		log.Fatalf(err.Error())
	}
	startTime := time.Now()
	reply, err := client.Exchange(packet, dst, nil)
	dt := time.Now().Sub(startTime)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if reply == nil {
		log.Println("No response from server")
		os.Exit(0)
	}

	if *verboseReply {
		fmt.Println("Auth-Type:", radCode{reply.Code})
		for _, attr := range reply.Attributes {
			if attrName, ok := reply.Dictionary.NameVID(attr.Vendor, attr.Type); ok {
				fmt.Println(attrName, ":", attr.Value)
			} else {
				log.Print("Err:", attr)
			}
		}
	} else if *verboseJSON {
		of := RadField{Code: radCode{reply.Code}}
		of.Attrs = make(map[string]string)
		for _, attr := range reply.Attributes {
			if attrName, ok := reply.Dictionary.NameVID(attr.Vendor, attr.Type); ok {
				of.Attrs[attrName] = fmt.Sprint(attr.Value)
			} else {
				log.Print("Err:", attr)
			}
		}
		d, _ := json.Marshal(of)
		fmt.Println(string(d))
	}
	if *qt {
		log.Print("Query time:", dt)
	}
}
