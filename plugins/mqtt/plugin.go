package mqtt

import (
	"fmt"
	"net/url"
	"sync"

	mqttpkg "github.com/iotaledger/goshimmer/packages/mqtt"
	"github.com/iotaledger/goshimmer/packages/shutdown"
	"github.com/iotaledger/goshimmer/packages/tangle"
	"github.com/iotaledger/goshimmer/plugins/config"
	"github.com/iotaledger/goshimmer/plugins/messagelayer"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"github.com/iotaledger/hive.go/workerpool"
	"github.com/labstack/echo/middleware"
)

// PluginName is the name of the graceful shutdown plugin.
const PluginName = "mqtt"

const (
	// RouteMQTT is the route for accessing the MQTT over WebSockets.
	RouteMQTT = "/mqtt"

	workerCount     = 1
	workerQueueSize = 10000
)

var (
	// plugin is the plugin instance of the issuer plugin.
	plugin *node.Plugin
	once   sync.Once
	log    *logger.Logger

	messageWorkerPool *workerpool.WorkerPool

	mqttBroker *mqttpkg.Broker
)

// Plugin gets the plugin instance.
func Plugin() *node.Plugin {
	once.Do(func() {
		plugin = node.NewPlugin(PluginName, node.Enabled, configure, run)
	})
	return plugin
}

func configure(*node.Plugin) {
	log = logger.NewLogger(PluginName)

	messageWorkerPool = workerpool.New(func(task workerpool.Task) {
		publishMessage(task.Param(0).(*tangle.CachedMessageEvent)) // metadata pass +1
		task.Return(nil)
	}, workerpool.WorkerCount(workerCount), workerpool.QueueSize(workerQueueSize), workerpool.FlushTasksAtShutdown(true))

	var err error
	mqttBroker, err = mqttpkg.NewBroker(config.Node().String(CfgMQTTBindAddress), config.Node().Int(CfgMQTTWSPort), "/ws", func(topic []byte) {
		log.Infof("Subscribe to topic: %s", string(topic))
	}, func(topic []byte) {
		log.Infof("Unsubscribe from topic: %s", string(topic))
	})

	if err != nil {
		log.Fatalf("MQTT broker init failed! %s", err)
	}

	setupWebSocketRoute()
}

func run(*node.Plugin) {

	if err := daemon.BackgroundWorker("MQTT Broker", func(shutdownSignal <-chan struct{}) {
		go func() {
			mqttBroker.Start()
			log.Infof("Starting MQTT Broker (port %s) ... done", mqttBroker.GetConfig().Port)
		}()

		if mqttBroker.GetConfig().Port != "" {
			log.Infof("You can now listen to MQTT via: http://%s:%s", mqttBroker.GetConfig().Host, mqttBroker.GetConfig().Port)
		}

		if mqttBroker.GetConfig().TlsPort != "" {
			log.Infof("You can now listen to MQTT via: https://%s:%s", mqttBroker.GetConfig().TlsHost, mqttBroker.GetConfig().TlsPort)
		}

		<-shutdownSignal
		log.Info("Stopping MQTT Broker ...")
		log.Info("Stopping MQTT Broker ... done")
	}, shutdown.PriorityMetrics); err != nil {
		log.Panicf("Failed to start as daemon: %s", err)
	}

	onMessageSolid := events.NewClosure(func(cachedMsgEvent *tangle.CachedMessageEvent) {
		if _, added := messageWorkerPool.TrySubmit(cachedMsgEvent); added {
			return // Avoid Release (done inside workerpool task)
		}
		cachedMsgEvent.MessageMetadata.Release()
		cachedMsgEvent.Message.Release()
	})

	if err := daemon.BackgroundWorker("MQTT Events", func(shutdownSignal <-chan struct{}) {
		log.Info("Starting MQTT Events ... done")

		messagelayer.Tangle().Events.MessageSolid.Attach(onMessageSolid)

		messageWorkerPool.Start()

		<-shutdownSignal
		log.Info("Stopping MQTT Events ...")

		messagelayer.Tangle().Events.MessageSolid.Detach(onMessageSolid)

		messageWorkerPool.StopAndWait()

		log.Info("Stopping MQTT Events ... done")
	}, shutdown.PriorityMetrics); err != nil {
		log.Panicf("Failed to start as daemon: %s", err)
	}

}

func setupWebSocketRoute() {

	// Configure MQTT WebSocket route
	mqttWSUrl, err := url.Parse(fmt.Sprintf("http://%s:%s", mqttBroker.GetConfig().Host, mqttBroker.GetConfig().WsPort))
	if err != nil {
		log.Fatalf("MQTT WebSocket init failed! %s", err)
	}

	wsGroup := webapi.Server().Group(RouteMQTT)
	proxyConfig := middleware.ProxyConfig{
		Skipper: middleware.DefaultSkipper,
		Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{
			{
				URL: mqttWSUrl,
			},
		}),
		// We need to forward any calls to the MQTT route to the ws endpoint of our broker
		Rewrite: map[string]string{
			RouteMQTT: mqttBroker.GetConfig().WsPath,
		},
	}

	wsGroup.Use(middleware.ProxyWithConfig(proxyConfig))
}
