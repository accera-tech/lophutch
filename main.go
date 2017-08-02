package main

import (
	"log"

	"github.com/spf13/viper"
	"github.com/tradeforce/lophutch/common"
	"github.com/tradeforce/lophutch/hutch"
)

func init() {
	if err := common.ConfigFlags(); err != nil {
		log.Fatalf("Error:\n%+v", err)
	}
}

var quit = make(chan struct{})

func main() {
	if viper.GetBool("run-once") {
		if err := hutch.Scout(); err != nil {
			log.Fatalf("Error:\n%+v", err)
		}
	} else {
		if err := hutch.Schedule(quit); err != nil {
			log.Fatalf("Error:\n%+v", err)
		}
	}
}
