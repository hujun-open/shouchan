![shouchan](./logo128.png)

[![CI](https://github.com/hujun-open/shouchan/actions/workflows/main.yml/badge.svg)](https://github.com/hujun-open/shouchan/actions/workflows/main.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hujun-open/shouchan)](https://pkg.go.dev/github.com/hujun-open/shouchan)

## Overview
Package shouchan provides simple configuration management for Golang application, with following features:

  - read configuration from command line flag and/or YAML file, mix&match, into a struct
  - 3 sources: default value, YAML file and flag
  - priority:in case of multiple source returns same config struct field, the preference is flag over YAML over default value
  - support following use cases:
	- simple: single action & single configuration, via SConf
	- multiple: multiple actions & multiple configurations, via ActionConf

include support following field types:
- integer types
- float type
- string
- time.Time
- time.Duration
- net.IP
- net.IPNet
- net.HardwareAddr (MAC addr)
- and all types that implement `encoding.TextMarshaler` and `encoding.TextUnmarshaler` interface

Additional types could be supported by using `Register`, see [types.go](types.go) for example. 

struct field naming:
  
  - YAML: lowercase
  - flag: lowercase, for nested struct field, it is "-<struct_fieldname_level_0>-<struct_fieldname_level_1>...", see example below:



## Simple Example:
```
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/hujun-open/shouchan"
)

type Company struct {
	//the usage tag is used for command line usage
	Name string `usage:"company name"`
}

type Employee struct {
	Name      string           `usage:"employee name"`
	Addr      string           `usage:"employee address"`
	IPAddr    net.IP           `usage:"employee IP address"`
	Subnet    net.IPNet        `usage:"employee IP subnet"`
	MAC       net.HardwareAddr `usage:"employee MAC address"`
	JointTime time.Time        `usage:"employee join time"`

	Employer Company
}

func main() {
	//default config
	def := Employee{
		Name:   "defName",
		Addr:   "defAddr",
		IPAddr: net.ParseIP("1.2.3.4"),
		MAC:    net.HardwareAddr{0x11, 0x22, 0x33, 0x44, 0x55, 0x66},
		Employer: Company{
			Name: "defCom",
		},
	}
	_, prefix, _ := net.ParseCIDR("192.168.1.0/24")
	def.Subnet = *prefix
	def.JointTime, _ = time.Parse(time.DateTime, "2023-01-02 13:22:33")
	cnf, err := shouchan.NewSConfCMDLine(&def, "")
	if err != nil {
		panic(err)
	}
	ferr, aerr := cnf.ReadwithCMDLine()
	fmt.Printf("ferr %v,aerr %v\n", ferr, aerr)
	fmt.Printf("final result is %+v\n", cnf.GetConf())
}
```
Output:

- Usage
```	
 .\test.exe -?
flag provided but not defined: -?
Usage:
  -f <filepath> : read from config file <filepath>
  -addr <string> : employee address
        default:defAddr
  -employer-name <string> : company name
        default:defCom
  -ipaddr <struct> : employee IP address
        default:1.2.3.4
  -jointtime <struct> : employee join time
        default:2023-01-02 13:22:33 +0000 UTC
  -mac <struct> : employee MAC address
        default:11:22:33:44:55:66
  -name <string> : employee name
        default:defName
  -subnet <struct> : employee IP subnet
        default:192.168.1.0/24
```    

- no command line args, no config file, default is used
```
 .\test.exe   
ferr <nil>,aerr <nil>
final result is &{Name:defName Addr:defAddr IPAddr:1.2.3.4 Subnet:{IP:192.168.1.0 Mask:ffffff00} MAC:11:22:33:44:55:66 JointTime:2023-01-02 13:22:33 +0000 UTC Employer:{Name:defCom}}
```

- config file via "-f" command args, value from file take procedence
```
.\test.exe -f ..\..\testdata\test.yaml
ferr <nil>,aerr <nil>
final result is &{Name:nameFromFile Addr:addrFromFile IPAddr:1.2.3.4 Subnet:{IP:192.168.1.0 Mask:ffffff00} MAC:11:22:33:44:55:66 JointTime:2023-01-02 13:22:33 +0000 UTC Employer:{Name:comFromFile}}
```
- mix command line args and config file, args to override employee name:
```
 .\test.exe -f ..\..\testdata\test.yaml -name nameFromArg
ferr <nil>,aerr <nil>
final result is &{Name:nameFromArg Addr:addrFromFile IPAddr:1.2.3.4 Subnet:{IP:192.168.1.0 Mask:ffffff00} MAC:11:22:33:44:55:66 JointTime:2023-01-02 13:22:33 +0000 UTC Employer:{Name:comFromFile}}
```
- mix command line args and config file, args to override company name:
```
.\test.exe -f ..\..\testdata\test.yaml -employer-name comFromArg
ferr <nil>,aerr <nil>
final result is &{Name:nameFromFile Addr:addrFromFile IPAddr:1.2.3.4 Subnet:{IP:192.168.1.0 Mask:ffffff00} MAC:11:22:33:44:55:66 JointTime:2023-01-02 13:22:33 +0000 UTC Employer:{Name:comFromArg}}
```


## Multiple Actions/Configurations Example

```
/*
This example demonstrate an zip file utility has two actions: show and zip, each action has different configuration via struct zipShowCfg and zipArchiveCfg.
*/
package main

import (
	"fmt"

	"github.com/hujun-open/shouchan"
)

type zipShowCfg struct {
	Zipf, Fname string
}

func (zscfg *zipShowCfg) Default() *zipShowCfg {
	return &zipShowCfg{
		Zipf:  "defaultShowZipf",
		Fname: "defaultShowFname",
	}
}

func (zacfg zipShowCfg) DefaultCfgFilePath() string {
	return ""
}

type zipArchiveCfg struct {
	Folder, Zipf string
}

func (zacfg *zipArchiveCfg) Default() *zipArchiveCfg {
	return &zipArchiveCfg{
		Zipf:   "defaultArchiveZipf",
		Folder: "defaultArchiveFolder",
	}
}

func (zacfg zipArchiveCfg) DefaultCfgFilePath() string {
	return ""
}

func main() {
	//create an ActionConf with a map 
	acnf, err := shouchan.NewActionConfWithCMDLine(map[string]shouchan.ConfigWithDefCfgFilePath{
		"show": (&zipShowCfg{}).Default(),
		"zip":  (&zipArchiveCfg{}).Default(),
	})
	if err != nil {
		panic(err)
	}
	err, ferr, aerr := acnf.ReadwithCMDLine()
	if err!=nil {
		panic(err)
	}
	fmt.Println("loading errors:",ferr, aerr)
	fmt.Printf("loaded action %v config is %+v", acnf.GetLoadedAction(), acnf.GetLoadedConf())
}

```
the command line looks like following:
```
Usage: <action> [<parameters...>]
Actions: show|zip
Action specific usage:
= show
Usage:
  -f <filepath> : read from config file <filepath>
  -fname <string> :
        default:defaultShowFname
  -zipf <string> :
        default:defaultShowZipf

= zip
Usage:
  -f <filepath> : read from config file <filepath>
  -folder <string> :
        default:defaultArchiveFolder
  -zipf <string> :
        default:defaultArchiveZipf
```
some CLI runs:
```
test.exe zip -folder testmy -zipf testmy.zip
<nil> <nil> <nil>
loaded action zip config is &{Folder:testmy Zipf:testmy.zip}
```

```
test.exe wrong
panic: wrong is not a valid action, use -? for list of actions

goroutine 1 [running]:
main.main()
        C:/hujun/gomodules/src/shouchan/cmd/test/main.go:50 +0x365
```