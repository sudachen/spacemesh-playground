package fake

import (
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/timesync"
	"sync"
	"time"
)

type TickProvider interface {
	Subscribe() timesync.LayerTimer
	Unsubscribe(timer timesync.LayerTimer)
	GetCurrentLayer() types.LayerID
	StartNotifying()
	GetGenesisTime() time.Time
	LayerToTime(id types.LayerID) time.Time
	Close()
	AwaitLayer(layerID types.LayerID) chan struct{}
}

// ManualClock is a clock that releases ticks on demand and not according to a real world clock
type manualClock struct {
	subs          map[timesync.LayerTimer]struct{}
	layerChannels map[types.LayerID]chan struct{}
	m             sync.Mutex
	currentLayer  types.LayerID
	genesisTime   time.Time
}

// LayerToTime returns the time of the provided layer
func (clk *manualClock) LayerToTime(types.LayerID) time.Time {
	return time.Now().Add(1000 * time.Hour) // hack so this wont take affect in the mock
}

// NewManualClock creates a new manual clock struct
func newManualClock(genesisTime time.Time) *manualClock {
	t := &manualClock{
		subs:          make(map[timesync.LayerTimer]struct{}),
		layerChannels: make(map[types.LayerID]chan struct{}),
		currentLayer:  0, // genesis
		genesisTime:   genesisTime,
	}
	return t
}

// Unsubscribe removes this channel ch from channels notified on tick
func (clk *manualClock) Unsubscribe(ch timesync.LayerTimer) {
	clk.m.Lock()
	delete(clk.subs, ch)
	clk.m.Unlock()
}

// StartNotifying is empty because this clock is manual
func (clk *manualClock) StartNotifying() {
}

var closedChannel chan struct{}

func init() {
	closedChannel = make(chan struct{})
	close(closedChannel)
}

// AwaitLayer implement the ability to notify a subscriber when a layer has ticked
func (clk *manualClock) AwaitLayer(layerID types.LayerID) chan struct{} {
	clk.m.Lock()
	defer clk.m.Unlock()
	if layerID <= clk.currentLayer {
		return closedChannel
	}
	if ch, found := clk.layerChannels[layerID]; found {
		return ch
	}
	ch := make(chan struct{})
	clk.layerChannels[layerID] = ch
	return ch
}

// Subscribe allow subscribes to be notified when a layer ticks
func (clk *manualClock) Subscribe() timesync.LayerTimer {
	ch := make(timesync.LayerTimer)
	clk.m.Lock()
	clk.subs[ch] = struct{}{}
	clk.m.Unlock()
	return ch
}

// Tick notifies all subscribers to this clock
func (clk *manualClock) Tick() {
	clk.m.Lock()
	defer clk.m.Unlock()

	clk.currentLayer++
	if ch, found := clk.layerChannels[clk.currentLayer]; found {
		close(ch)
		delete(clk.layerChannels, clk.currentLayer)
	}
	for s := range clk.subs {
		s <- clk.currentLayer
	}
}

// GetCurrentLayer gets the last ticked layer
func (clk *manualClock) GetCurrentLayer() types.LayerID {
	clk.m.Lock()
	defer clk.m.Unlock()

	return clk.currentLayer
}

// GetGenesisTime returns the set genesis time for this clock
func (clk *manualClock) GetGenesisTime() time.Time {
	clk.m.Lock()
	defer clk.m.Unlock()
	return clk.genesisTime
}

// Close does nothing because this clock is manual
func (clk *manualClock) Close() {}

