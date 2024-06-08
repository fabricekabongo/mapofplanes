package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "geo_db_grid_locations_total",
		Help: "The total number of locations in the grid",
	}, []string{"grid_name"})
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "geo_db_grid_processed_ops_total",
		Help: "The total number of processed operations by the grid",
	})
)

type LocationUpdateEvent struct {
	LocId   string  `json:"loc_id"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	PrevLat float64 `json:"prev_lat"`
	PrevLon float64 `json:"prev_lon"`
}

type LocationAddedEvent struct {
	LocId string  `json:"loc_id"`
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
}

type LocationDeletedEvent struct {
	LocId string `json:"loc_id"`
}

type Grid struct {
	Name                   string
	locations              map[string]*LocationEntity
	addEventSubscribers    map[string]chan LocationAddedEvent
	updateEventSubscribers map[string]chan LocationUpdateEvent
	deleteEventSubscribers map[string]chan LocationDeletedEvent
	gauge                  prometheus.Gauge
}

func NewGrid(name string) *Grid {

	return &Grid{
		Name:                   name,
		locations:              make(map[string]*LocationEntity),
		addEventSubscribers:    make(map[string]chan LocationAddedEvent),
		updateEventSubscribers: make(map[string]chan LocationUpdateEvent),
		deleteEventSubscribers: make(map[string]chan LocationDeletedEvent),
		gauge:                  metricGauge.WithLabelValues(name),
	}
}

func (g *Grid) DeleteLocation(loc *LocationEntity) {
	defer g.gauge.Dec()
	defer opsProcessed.Inc()

	if len(loc.LocId) == 0 {
		panic("locationId is required. It should have never reached the grid")
	}
	_, ok := g.locations[loc.LocId]

	if !ok {
		return
	}

	delete(g.locations, loc.LocId)

	go func() {
		for _, subscriber := range g.deleteEventSubscribers {
			subscriber <- LocationDeletedEvent{LocId: loc.LocId}
		}

	}()
}

func (g *Grid) UpdateLocation(loc *LocationEntity, lat float64, lon float64) error {
	defer opsProcessed.Inc()

	_, ok := g.locations[loc.LocId]
	if !ok {
		panic("Location is not in the grid. Update should have never reached the grid")
	}
	prevLat := g.locations[loc.LocId].Lat
	prevLon := g.locations[loc.LocId].Lon

	g.locations[loc.LocId].Lat = lat
	g.locations[loc.LocId].Lon = lon

	go func() {
		for _, subscriber := range g.updateEventSubscribers {
			subscriber <- LocationUpdateEvent{
				LocId:   loc.LocId,
				Lat:     lat,
				Lon:     lon,
				PrevLat: prevLat,
				PrevLon: prevLon,
			}
		}
	}()

	return nil
}

func (g *Grid) AddLocation(loc *LocationEntity) {
	defer opsProcessed.Inc()
	defer g.gauge.Inc()
	_, ok := g.locations[loc.LocId]
	if ok {
		panic("Location already exists in the grid. Add should have never reached the grid")
	}

	go func() {
		for _, subscriber := range g.addEventSubscribers {
			subscriber <- LocationAddedEvent{
				LocId: loc.LocId,
				Lat:   loc.Lat,
				Lon:   loc.Lon,
			}
		}
	}()
}
