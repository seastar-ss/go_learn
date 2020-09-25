package main

import "fmt"
//import "strconv"

type IPAddr [4]byte

// TODO: 给 IPAddr 添加一个 "String() string" 方法

func (ip *IPAddr) String() string{
	fmt.Println("run ")
	return fmt.Sprintf("%v.%v.%v.%v",ip[0],ip[1],ip[2],ip[3])
}

func main() {
	
	hosts := map[string]IPAddr{
		"loopback":  {127, 0, 0, 1},
		"googleDNS": {8, 8, 8, 8},
	}
	//fmt.Printf("test:%s",hosts['loopback'])
	for name, ip := range hosts {
		fmt.Printf("%v: %v\n", name, &ip)
		//fmt.Printf("%v: %v\n", name, ip.String())
	}
}