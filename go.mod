module oes-acceptance-testing/v1

go 1.14

require (
	github.com/confluentinc/confluent-kafka-go v1.3.0
	github.com/dustin/go-coap v0.0.0-20190908170653-752e0f79981e
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/mitchellh/mapstructure v1.3.2 // indirect
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/rs/xid v1.2.1
	github.com/sigurn/crc16 v0.0.0-20160107003519-da416fad5162
	github.com/sigurn/utils v0.0.0-20190728110027-e1fefb11a144 // indirect
	github.com/spf13/afero v1.3.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.0
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8 // indirect
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	golang.org/x/text v0.3.3 // indirect
	gopkg.in/ini.v1 v1.57.0 // indirect
)

replace github.com/dustin/go-coap v0.0.0-20190908170653-752e0f79981e => ./pkg/go-coap
