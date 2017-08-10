package main

import (
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/tradeforce/lophutch/common"
	"github.com/tradeforce/lophutch/hutch"
)

func init() {
	if err := common.ConfigFlags(); err != nil {
		log.Fatalf("Error:\n%+v", err)
	}
}

var done = make(chan struct{})

func main() {
	if viper.GetBool("run-once") {
		delays := make(map[string]time.Time)
		if err := hutch.Scout(delays); err != nil {
			log.Fatalf("Error:\n%+v", err)
		}
	} else {
		if err := hutch.Schedule(done); err != nil {
			log.Fatalf("Error:\n%+v", err)
		}
	}
}
