package main

import (
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"iotdb-courier/emitter/internal/config"
	"iotdb-courier/emitter/internal/listener"
)

var configFile = flag.String("f", "etc/emitter-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	lsnr := listener.NewEmitterListener(c)
	defer lsnr.Stop()

	fmt.Printf("Starting listening at %s(%s) \n", c.UsingQueue, c.UsingTopic)
	lsnr.Start()

}
