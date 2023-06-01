package env

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"log"
)

var LubanFork = uint64(29295050)

func init() {
	viper.SetEnvPrefix("bsc")
	bindAndGet := func(key string) interface{} {
		if err := viper.BindEnv(key); err != nil {
			panic(err)
		}
		return viper.Get(key)
	}
	if v := bindAndGet("luban_fork"); v != nil {
		log.Printf("change : luban_fork=%s\n", v)
		LubanFork = cast.ToUint64(v)
	}
}
