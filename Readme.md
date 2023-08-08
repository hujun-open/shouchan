![shouchan](./logo128.png)

[![CI](https://github.com/hujun-open/shouchan/actions/workflows/main.yml/badge.svg)](https://github.com/hujun-open/shouchan/actions/workflows/main.yml)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hujun-open/shouchan)](https://pkg.go.dev/github.com/hujun-open/shouchan)

## Overview
Package shouchan provides simple configuration management for Golang application, with following features:

  - read configuration from command line flag and/or YAML file, mix&match, into a struct
  - 3 sources: default value, YAML file and flag
  - priority:in case of multiple source returns same config struct field, the preference is flag over YAML over default value

include support following field types:
- integer types
- float type
- string
- time.Duration
- and all types that implement `encoding.TextMarshaler` and `encoding.TextUnmarshaler` interface
- there is also `github.com/hujun-open/shouchantypes` include some other types

Additional types could be supported by using `Register`, see [github.com/hujun-open/shouchantypes](https://github.com/hujun-open/shouchantypes) for example. 

## CLI & YAML Support

- YAML: shouchan uses [extyaml](https://pkg.go.dev/github.com/hujun-open/extyaml) for YAML marshal and unmarshal 
- CLI Flag: shouchan uses [myflags](https://pkg.go.dev/github.com/hujun-open/myflags) for command line flag generation

refer to corresponding doc for details on CLI & YAML support. 


## Example:
https://github.com/hujun-open/shouchan/blob/809751e636ae1230134d11983523ec5d8a2b24e6/example/main.go#L1-L57

Output:

- Usage
```	
 .\test.exe -?
flag provided but not defined: -?
shouchan example
  - addr: employee address
        default:defAddrPointer
  - employer-name: company name
        default:defCom
  - ipaddr: employee IP address
        default:1.2.3.4
  - jointtime: employee join time
        default:2023-01-02 13:22:33
  - mac: employee MAC address
        default:11:22:33:44:55:66
  - n2addr:
        default:1.1.1.1
  - naddr:
        default:2.2.2.2
  - name: employee name
        default:defName
  - subnet: employee IP subnet
        default:192.168.1.0/24

  -cfgfromfile: load configuration from the specified file
        default:test.yaml
```    

- no command line args, no config file, default is used
```
 .\test.exe   
ferr failed to open config file test.yaml, open test.yaml: The system cannot find the file specified.,aerr <nil>
final result is &{Name:defName Addr:0xc0000528b0 Naddr:2.2.2.2 N2addr:1.1.1.1 IPAddr:1.2.3.4 Subnet:{IP:192.168.1.0 Mask:ffffff00} MAC:11:22:33:44:55:66 JointTime:2023-01-02 13:22:33 +0000 UTC Employer:{Name:defCom}}
```

- config file via "-f" command args, value from file take procedence
```
 .\test.exe -cfgfromfile cfg.yaml
ferr <nil>,aerr <nil>
final result is &{Name:nameFromFile Addr:0xc0000528b0 Naddr:2.2.2.2 N2addr:1.1.1.1 IPAddr:1.2.3.4 Subnet:{IP:192.168.1.0 Mask:ffffff00} MAC:11:22:33:44:55:66 JointTime:2023-01-02 13:22:33 +0000 UTC Employer:{Name:comFromFile}}
```
- mix command line args and config file, args to override employee name:
```
.\test.exe -cfgfromfile cfg.yaml -name nameFromArg
ferr <nil>,aerr <nil>
final result is &{Name:nameFromArg Addr:0xc000088880 Naddr:2.2.2.2 N2addr:1.1.1.1 IPAddr:1.2.3.4 Subnet:{IP:192.168.1.0 Mask:ffffff00} MAC:11:22:33:44:55:66 JointTime:2023-01-02 13:22:33 +0000 UTC Employer:{Name:comFromFile}}
```
- mix command line args and config file, args to override company name:
```
.\test.exe -cfgfromfile cfg.yaml -employer-name comFromArg
ferr <nil>,aerr <nil>
final result is &{Name:nameFromFile Addr:0xc000104880 Naddr:2.2.2.2 N2addr:1.1.1.1 IPAddr:1.2.3.4 Subnet:{IP:192.168.1.0 Mask:ffffff00} MAC:11:22:33:44:55:66 JointTime:2023-01-02 13:22:33 +0000 UTC Employer:{Name:comFromArg}}
```

## Code Generation
shouchan also provides a code generation tool to deal with large amount of constants, see [shouchangen](https://github.com/hujun-open/shouchangen).
