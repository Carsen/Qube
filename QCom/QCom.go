package QCom

import (
	"log"
	"net"
)

func GetInterfaces() {
	intf, err := net.Interfaces()

	if err != nil {
		log.Fatal(err)
	}

	return intf
}
