package QCom

import (
	"log"
	"net"
)

func IfaceAmt() int{
	intfc, err := net.Interfaces()
	if err != nil{
		log.Fatal(err)
	}

	for _, Iface := range intfc {
//		addrs, err := Iface.Addrs()
//		if err != nil {
//			log.Fatal(err)
//		}
		return Iface.Index
	}

	return 0
}
