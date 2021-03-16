package autopeeringanalysis

import (
	"fmt"
	"time"
)

var (
	fileName = "autopeering-analysis.csv"
)

// TableDescription holds the description of the autopeering analysis.
var TableDescription = []string{
	"Time",
	"KnownPeers",
	"InboundNeighbors",
	"OutboundNeighbors",
}

// AutopeeringInfo holds the information of the autopeering.
type AutopeeringInfo struct {
	Time              time.Time
	KnownPeers        int
	InboundNeighbors  int
	OutboundNeighbors int
}

func (a AutopeeringInfo) toCSV() (row []string) {
	row = append(row, []string{
		fmt.Sprint(a.Time.UnixNano()),
		fmt.Sprint(a.KnownPeers),
		fmt.Sprint(a.InboundNeighbors),
		fmt.Sprint(a.OutboundNeighbors)}...)

	return
}
