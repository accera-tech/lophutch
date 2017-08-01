package main

import (
	"log"

	"github.com/tradeforce/lophutch/common"
)

func init() {
	if err := common.ConfigFlags(); err != nil {
		log.Fatalf("Error:\n%+v", err)
	}
}

func main() {

}
