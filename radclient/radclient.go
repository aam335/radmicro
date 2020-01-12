package main

import (
	"bufio"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"time"

	_ "flag"

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
	log.Fatal(s, ":", err)
}

func main() {
	// log.Println(os.Args)
	var serverIPPort = flag.String("serv", "127.0.0.1:1812", "Radius adress:port")
	var localIPPort = flag.String("me", "", "Source adress:port")
	var timeoutStr = flag.String("t", "1s", "timeout")
	var retries = flag.Int("r", 1, "retries")
	var secret = flag.String("secret", "", "Radius secret")

	flag.Parse()

	if *secret == "" {
		usage("secret", errors.New("must be set"))
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

	packet := radius.New(radius.CodeAccessRequest, []byte(*secret))
	// packet.Add("Calling-Station-Id", "NAS-Fake")
	reader := bufio.NewReader(os.Stdin)
	rx := regexp.MustCompile("\\s*:\\s*")
	for {
		strAttr, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		av := rx.Split(strAttr, 2)
		if len(av) != 2 {
			log.Fatalf("'%v' not in attr:val format", strAttr)
		}
		if err := packet.Add(av[0], av[1]); err != nil {
			log.Fatalf("'%v' error:", strAttr)
		}
	}

	dst, err := ParceUDPAddr(*serverIPPort)
	if err != nil {
		log.Fatalf(err.Error())
	}

	reply, err := client.Exchange(packet, dst, nil)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if reply == nil {
		log.Println("No response from server")
		os.Exit(0)
	}

	switch reply.Code {
	case radius.CodeAccessAccept:
		log.Println("Accept")
	case radius.CodeAccessReject:
		log.Println("Reject")
	}
}
