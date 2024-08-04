package QCom

import (
	"log"
	"net"
	"strconv"
)

func IfaceAmt() int {
	intfc, err := net.Interfaces()
	if err != nil {
		log.Fatal(err)
	}

	for _, Iface := range intfc {
		//		addrs, err := Iface.Addrs()
		//		if err != nil {
		//			log.Fatal(err)
		//		}
		i := strconv.Itoa(Iface.Index)
		return i
	}

	return 0
}
