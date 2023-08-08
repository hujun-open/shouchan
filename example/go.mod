module test

go 1.20

require (
	github.com/hujun-open/shouchan v0.2.0
	github.com/hujun-open/shouchantypes v0.0.0-00010101000000-000000000000
)

require (
	github.com/hujun-open/extyaml v0.4.0 // indirect
	github.com/hujun-open/myflags v0.3.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/hujun-open/shouchan => ../

replace github.com/hujun-open/shouchantypes => ../../shouchantypes
