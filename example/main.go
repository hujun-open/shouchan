package main

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/hujun-open/myflags/v2"
	"github.com/hujun-open/shouchan/v2"
	_ "github.com/hujun-open/shouchantypes/v2"
)

type Company struct {
	//the usage tag is used for command line usage
	Name string `usage:"company name"`
}

type Employee struct {
	Name      string  `usage:"employee name"`
	Addr      *string `usage:"employee address"`
	Naddr     *netip.Addr
	N2addr    netip.Addr
	IPAddr    net.IP           `usage:"employee IP address"`
	Subnet    net.IPNet        `usage:"employee IP subnet"`
	MAC       net.HardwareAddr `usage:"employee MAC address"`
	JointTime time.Time        `usage:"employee join time"`
	IsRetired bool             `alias:"r" usage:"retired"`

	Employer Company
}

func main() {
	//default config
	def := Employee{
		Name:   "defName",
		N2addr: netip.AddrFrom4([4]byte{1, 1, 1, 1}),
		IPAddr: net.ParseIP("1.2.3.4"),
		MAC:    net.HardwareAddr{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
		Employer: Company{
			Name: "defCom",
		},
	}
	def.Addr = new(string)
	*def.Addr = "defAddrPointer"
	def.Naddr = new(netip.Addr)
	*def.Naddr = netip.AddrFrom4([4]byte{2, 2, 2, 2})
	_, prefix, _ := net.ParseCIDR("192.168.1.0/24")
	def.Subnet = *prefix
	def.JointTime, _ = time.Parse(time.DateTime, "2023-01-02 13:22:33")
	cnf, err := shouchan.NewSConf(&def, "example", "shouchan example",
		shouchan.WithDefaultConfigFilePath[Employee]("test.yaml"),
		shouchan.WithFillOptions[Employee]([]myflags.FillerOption{
			myflags.WithRootMethod(myflags.DefRunMethod)}),
	)
	if err != nil {
		panic(err)
	}

	cmd, ferr, aerr := cnf.ReadwithCMDLine()
	fmt.Printf("command get executed is %v\n", cmd.Name())
	fmt.Printf("ferr %v,aerr %v\n", ferr, aerr)
	fmt.Printf("final result is %+v\n", cnf.GetConf())

}
