package autopeeringanalysis

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/shutdown"
	"github.com/iotaledger/goshimmer/plugins/autopeering"
	"github.com/iotaledger/hive.go/autopeering/discover"
	"github.com/iotaledger/hive.go/autopeering/selection"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

// PluginName is the name of the autopeering analysis plugin.
const PluginName = "AutopeeringAnalysis"

var (
	// plugin is the plugin instance of the autopeering plugin.
	plugin      *node.Plugin
	once        sync.Once
	apInfo      AutopeeringInfo
	apInfoMutex sync.Mutex

	log *logger.Logger
	f   *os.File
	w   *csv.Writer
)

// Plugin gets the plugin instance.
func Plugin() *node.Plugin {
	once.Do(func() {
		plugin = node.NewPlugin(PluginName, node.Disabled, configure, run)
	})
	return plugin
}

func init() {
	// If the file doesn't exist, create it, or truncate the file
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(filepath.Dir(ex), fileName)

	f, err = os.Create(path)
	if err != nil {
		panic(err)
	}

	w = csv.NewWriter(f)
	// write TableDescription
	if err = w.Write(TableDescription); err != nil {
		panic(err)
	}
}

func configure(*node.Plugin) {
	log = logger.NewLogger(PluginName)

	configureEvents()
}

func run(*node.Plugin) {
	if err := daemon.BackgroundWorker(PluginName, start, shutdown.PriorityGossip); err != nil {
		log.Panicf("Failed to start as daemon: %s", err)
	}

}

func start(shutdownSignal <-chan struct{}) {
	defer log.Info("Stopping " + PluginName + " ... done")
	<-shutdownSignal
	w.Flush()
	f.Close()
}

func configureEvents() error {
	// assure that the autopeering is instantiated
	peerDisc := autopeering.Discovery()
	peerSel := autopeering.Selection()

	// log the peer discovery events
	peerDisc.Events().PeerDiscovered.Attach(events.NewClosure(func(ev *discover.DiscoveredEvent) {
		apInfoMutex.Lock()
		defer apInfoMutex.Unlock()
		apInfo.Time = time.Now()
		apInfo.KnownPeers++
		if err := w.Write(apInfo.toCSV()); err != nil {
			log.Error(err)
		}

	}))
	peerDisc.Events().PeerDeleted.Attach(events.NewClosure(func(ev *discover.DeletedEvent) {
		apInfoMutex.Lock()
		defer apInfoMutex.Unlock()
		apInfo.Time = time.Now()
		apInfo.KnownPeers--
		if err := w.Write(apInfo.toCSV()); err != nil {
			log.Error(err)
		}
	}))

	peerSel.Events().OutgoingPeering.Attach(events.NewClosure(func(ev *selection.PeeringEvent) {
		if ev.Status {
			apInfoMutex.Lock()
			defer apInfoMutex.Unlock()

			apInfo.Time = time.Now()
			apInfo.Neighbors++
			if err := w.Write(apInfo.toCSV()); err != nil {
				log.Error(err)
			}
		}
	}))
	peerSel.Events().IncomingPeering.Attach(events.NewClosure(func(ev *selection.PeeringEvent) {
		if ev.Status {
			apInfoMutex.Lock()
			defer apInfoMutex.Unlock()

			apInfo.Time = time.Now()
			apInfo.Neighbors++
			if err := w.Write(apInfo.toCSV()); err != nil {
				log.Error(err)
			}
		}
	}))
	peerSel.Events().Dropped.Attach(events.NewClosure(func(ev *selection.DroppedEvent) {
		apInfoMutex.Lock()
		defer apInfoMutex.Unlock()
		apInfo.Time = time.Now()
		apInfo.Neighbors--
		if err := w.Write(apInfo.toCSV()); err != nil {
			log.Error(err)
		}
	}))
	return nil
}

// func getAutopeeringInfo() (a AutopeeringInfo) {
// 	a.Time = time.Now()
// 	a.KnownPeers = len(autopeering.Discovery().GetVerifiedPeers())
// 	a.Neighbors = len(autopeering.Selection().GetIncomingNeighbors())
// 	a.OutboundNeighbors = len(autopeering.Selection().GetOutgoingNeighbors())
// 	return a
// }
