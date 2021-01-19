package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/client"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	cfgNodeURI      = "node"
	cfgMessage      = "message"
	numberOfSpammer = int(1500)
	spamN           = "spamN"
)

func init() {
	flag.String(cfgNodeURI, "http://localhost:8080", "the URI of the node API")
	flag.String(cfgMessage, "Hello Tangle", "the URI of the node API")
	flag.Int(spamN, 1000, "number of spammer worker") // this does not work
}

func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	goshimAPI := client.NewGoShimmerAPI(viper.GetString(cfgNodeURI))
	messageBytes := []byte(viper.GetString(cfgMessage))
	var issued, failed int
	for {
		fmt.Printf("                                                      \r")
		fmt.Printf("issued %d, failed %d, worker %d\r", issued, failed, id)
		_, err := goshimAPI.Data(messageBytes)
		if err != nil {
			failed++
			continue
		}
		issued++
		time.Sleep(time.Duration(rand.ExpFloat64()) * time.Second)
	}
}

func main() {
	flag.Parse()
	if err := viper.BindPFlags(flag.CommandLine); err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	numberOfSpammer := viper.GetInt(spamN)
	for i := 1; i <= numberOfSpammer; i++ {
		wg.Add(1)
		go worker(i, &wg)
	}
	wg.Wait()
}
