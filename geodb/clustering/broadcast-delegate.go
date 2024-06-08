package clustering

import (
	"bytes"
	"encoding/gob"
	"fabricekabongo.com/geodb/world"
	"github.com/hashicorp/memberlist"
	"log"
)

type BroadcastDelegate struct {
	state      *NodeState
	broadcasts *memberlist.TransmitLimitedQueue
}

type NodeState struct {
	World *world.Map
}

func NewBroadcastDelegate(world *world.Map, broadcasts *memberlist.TransmitLimitedQueue) *BroadcastDelegate {
	return &BroadcastDelegate{
		state: &NodeState{
			World: world,
		},
		broadcasts: broadcasts,
	}
}

func (d *BroadcastDelegate) NodeMeta(limit int) []byte {
	return []byte{} // Metadata information
}

func (d *BroadcastDelegate) NotifyMsg(buf []byte) {
	if len(buf) > 0 {
		dec := gob.NewDecoder(bytes.NewReader(buf))
		// Process the message
		var location world.LocationEntity

		err := dec.Decode(&location)
		if err != nil {
			return
		}

		err = d.state.World.Save(location.LocId, location.Lat, location.Lon)
		if err != nil {
			return
		}
	}
}

func (d *BroadcastDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return d.broadcasts.GetBroadcasts(overhead, limit)
}

func (d *BroadcastDelegate) LocalState(join bool) []byte {
	if join {
		log.Println("Sharing local state to a new node")
	} else {
		log.Println("Sharing local state for routine sync")
	}

	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	err := enc.Encode(d.state.World)
	if err != nil {
		return []byte{}
	}

	return buf.Bytes()
}

func (d *BroadcastDelegate) MergeRemoteState(buf []byte, join bool) {
	if join {
		log.Println("Getting state from the cluster to start well")
	} else {
		log.Println("Getting state from the cluster for routine sync")
	}

	dec := gob.NewDecoder(bytes.NewReader(buf))
	var worldMap world.Map

	err := dec.Decode(&worldMap)
	if err != nil {
		return
	}

	go func(worldMap world.Map) {
		for _, loc := range worldMap.Locations {
			err := d.state.World.Save(loc.LocId, loc.Lat, loc.Lon)
			if err != nil {
				continue
			}
		}
	}(worldMap)
}
