package autopeeringanalysis

import (
	"encoding/csv"
	"os"
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/plugins/autopeering"
	"github.com/iotaledger/hive.go/autopeering/discover"
	"github.com/iotaledger/hive.go/autopeering/selection"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

// PluginName is the name of the autopeering analysis plugin.
const PluginName = "AutopeeringAnalysis"

var (
	// plugin is the plugin instance of the autopeering plugin.
	plugin *node.Plugin
	once   sync.Once

	log *logger.Logger
)

// Plugin gets the plugin instance.
func Plugin() *node.Plugin {
	once.Do(func() {
		plugin = node.NewPlugin(PluginName, node.Disabled, run)
	})
	return plugin
}

func run(*node.Plugin) {
	log = logger.NewLogger(PluginName)

	configureEvents(fileName)
}

func configureEvents(filePath string) error {
	// assure that the autopeering is instantiated
	peerDisc := autopeering.Discovery()
	peerSel := autopeering.Selection()

	// If the file doesn't exist, create it, or truncate the file
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)

	// write TableDescription
	if err := w.Write(TableDescription); err != nil {
		panic(err)
	}

	// log the peer discovery events
	peerDisc.Events().PeerDiscovered.Attach(events.NewClosure(func(ev *discover.DiscoveredEvent) {
		w.Write(getAutopeeringInfo().toCSV())
		w.Flush()
	}))
	// peerDisc.Events().PeerDeleted.Attach(events.NewClosure(func(ev *discover.DeletedEvent) {
	// 	log.Infof("Removed offline: %s / %s", ev.Peer.Address(), ev.Peer.ID())
	// }))

	peerSel.Events().OutgoingPeering.Attach(events.NewClosure(func(ev *selection.PeeringEvent) {
		if ev.Status {
			w.Write(getAutopeeringInfo().toCSV())
			w.Flush()
		}
	}))
	peerSel.Events().IncomingPeering.Attach(events.NewClosure(func(ev *selection.PeeringEvent) {
		if ev.Status {
			w.Write(getAutopeeringInfo().toCSV())
			w.Flush()
		}
	}))
	// peerSel.Events().Dropped.Attach(events.NewClosure(func(ev *selection.DroppedEvent) {
	// 	log.Infof("Peering dropped: %s", ev.DroppedID)
	// }))
	return nil
}

func getAutopeeringInfo() (a AutopeeringInfo) {
	a.Time = time.Now()
	a.KnownPeers = len(autopeering.Discovery().GetVerifiedPeers())
	a.InboundNeighbors = len(autopeering.Selection().GetIncomingNeighbors())
	a.OutboundNeighbors = len(autopeering.Selection().GetOutgoingNeighbors())
	return a
}
