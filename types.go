package shouchan

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register[net.HardwareAddr](macTtoStr, macFromStr)
	Register[net.IPNet](ipnetTtoStr, ipnetFromStr)
	Register[time.Time](timetoStr, timeFromStr)
	Register[net.IP](iptoStr, ipFromStr)
}

// support formats: xx:xx:xx:xx:xx:xx, xx-xx-xx-xx-xx-xx
func macFromStr(text string) (any, error) {
	if text == "" {
		return net.HardwareAddr{}, nil
	}
	var r = make([]byte, 6)
	var flist []string
	switch {
	case strings.Contains(text, "-"):
		flist = strings.Split(text, "-")
	case strings.Contains(text, ":"):
		flist = strings.Split(text, ":")
	default:
		return nil, fmt.Errorf("can't find supported MAC format")
	}
	for i, v := range flist {
		x, err := strconv.ParseInt(strings.TrimSpace(v), 16, 64)
		if err != nil {
			return nil, fmt.Errorf("%v is not valid byte value in hex", v)
		}
		if x >= 255 {
			return nil, fmt.Errorf("%v is not valid byte value in hex, should be <256", v)
		}
		r[i] = byte(x)
	}
	return net.HardwareAddr(r), nil

}

// just the net.Hardware.String()
func macTtoStr(in any) (string, error) {
	v := in.(net.HardwareAddr)
	return v.String(), nil
}

func ipnetFromStr(text string) (any, error) {
	_, r, err := net.ParseCIDR(text)
	return *r, err
}

func ipnetTtoStr(in any) (string, error) {
	v := in.(net.IPNet)
	return v.String(), nil
}

const TimeLayout = time.DateTime

func timeFromStr(text string) (any, error) {
	return time.Parse(TimeLayout, text)
}

func timetoStr(in any) (string, error) {
	return in.(time.Time).Format(TimeLayout), nil
}

func ipFromStr(text string) (any, error) {
	addr := net.ParseIP(text)
	if addr == nil {
		return nil, fmt.Errorf("%v is not a valid IP address", text)
	}
	return addr, nil
}

func iptoStr(in any) (string, error) {
	return in.(net.IP).String(), nil
}
