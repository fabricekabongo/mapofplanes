package main

import (
	"github.com/uber/h3-go"
	"log"
)

const (
	LocationChangeTypeAdd    = "add"
	LocationChangeTypeDelete = "delete"
	LocationChangeTypeUpdate = "update"
	ListenGridUpdate         = "listen"
)

type LocationEntity struct {
	LocationId string  `json:"locationId"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}

type LocationChangeEvent struct {
	Location   LocationEntity
	ChangeType string
}

type Grid struct {
	name        string
	Locations   map[string]*LocationEntity
	Subscribers map[string]chan LocationChangeEvent
}

func (g *Grid) DeleteLocation(locationId string) {
	delete(g.Locations, locationId)
	g.updateSubscribers(LocationEntity{LocationId: locationId}, LocationChangeTypeDelete)
}

func (g *Grid) UpdateLocation(location *LocationEntity) {
	g.Locations[location.LocationId] = location
	g.updateSubscribers(*location, LocationChangeTypeUpdate)
}

func (g *Grid) updateSubscribers(location LocationEntity, changeType string) {
	update := LocationChangeEvent{
		Location:   location,
		ChangeType: changeType,
	}

	for _, subscriber := range g.Subscribers {
		subscriber <- update
	}
}

func (g *Grid) AddLocation(location *LocationEntity) {
	oldLocation, ok := g.Locations[location.LocationId]
	if !ok {
		g.Locations[location.LocationId] = location
		g.updateSubscribers(*location, LocationChangeTypeAdd)
		return
	}

	if oldLocation.Latitude != location.Latitude || oldLocation.Longitude != location.Longitude {
		g.Locations[location.LocationId] = location
		g.updateSubscribers(*location, LocationChangeTypeUpdate)
	}
}

func (g *Grid) Subscribe(subscriberId string, subscriber chan LocationChangeEvent) {
	g.Subscribers[subscriberId] = subscriber
}

func (g *Grid) Unsubscribe(subscriberId string) {
	delete(g.Subscribers, subscriberId)
}

type Map struct {
	Locations map[string]*LocationEntity
	Grids     map[string]*Grid
}

// { "Command":"add", "LocationEntity":{"LocationId": "the-boss", "Latitude":"25.6", Longitude:"52.0"}}
func NewMap() *Map {
	return &Map{
		Locations: make(map[string]*LocationEntity),
		Grids:     make(map[string]*Grid),
	}
}

func (m *Map) Subscribe(gridName string, subscriber chan LocationChangeEvent) {
	grid, ok := m.Grids[gridName]
	if !ok {
		log.Println("Grid not found: ", gridName)
		return
	}

	grid.Subscribe(gridName, subscriber)
}

func (m *Map) AddLocation(location *LocationEntity) {
	oldLocation, ok := m.Locations[location.LocationId]
	if !ok {
		m.Locations[location.LocationId] = location
		grid := m.getGrip(location.Latitude, location.Longitude)
		grid.AddLocation(location)
		return
	}

	if oldLocation.Latitude != location.Latitude || oldLocation.Longitude != location.Longitude {
		oldGrid := m.getGrip(oldLocation.Latitude, oldLocation.Longitude)
		grid := m.getGrip(location.Latitude, location.Longitude)

		if oldGrid != grid {
			oldGrid.DeleteLocation(location.LocationId)
			grid.AddLocation(location)
			log.Println("Added location: ", location, " to grid: ", grid.name)
		} else {
			grid.UpdateLocation(location)
			log.Println("Updated location: ", location, " in grid: ", grid.name)
		}
	}

}

//86283082fffffff

func (m *Map) getGrip(lat float64, lon float64) *Grid {
	geo := h3.GeoCoord{
		Latitude:  lat,
		Longitude: lon,
	}

	geoHash := h3.FromGeo(geo, 6)
	geoHashString := h3.ToString(geoHash)

	grid, ok := m.Grids[geoHashString]
	if !ok {
		grid = &Grid{
			name:        geoHashString,
			Locations:   make(map[string]*LocationEntity),
			Subscribers: make(map[string]chan LocationChangeEvent),
		}
		m.Grids[geoHashString] = grid
	}

	return grid
}
