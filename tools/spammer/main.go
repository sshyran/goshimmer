package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/iotaledger/goshimmer/client"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cfgNodeURI = "node"
	cfgMessage = "message"
	mps        = "mps"
)

func init() {
	flag.String(cfgNodeURI, "http://localhost:8080", "the URI of the node API")
	flag.String(cfgMessage, "Hello Tangle", "the URI of the node API")
	flag.Int(mps, 10000, "mean maximal number of messages per second") // this does not work
}

func spam(mps int) {
	goshimAPI := client.NewGoShimmerAPI(viper.GetString(cfgNodeURI))
	messageBytes := []byte(viper.GetString(cfgMessage))
	var issued, failed int
	for {
		fmt.Printf("issued %d, failed %d\r", issued, failed)
		_, err := goshimAPI.Data(messageBytes)
		if err != nil {
			failed++
			continue
		}
		issued++
		time.Sleep(time.Duration(rand.ExpFloat64()/float64(mps)) * time.Second)
	}
}

func main() {
	flag.Parse()
	if err := viper.BindPFlags(flag.CommandLine); err != nil {
		panic(err)
	}
	spam(viper.GetInt(mps))
}
