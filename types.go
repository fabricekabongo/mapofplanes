package main

type TRANSMISSION_TYPE int8
type MESSAGE_TYPE string

const (
	ES_IDENT_AND_CATEGORY TRANSMISSION_TYPE = 1
	ES_SURFACE_POS        TRANSMISSION_TYPE = 2
	ES_AIRBORNE_POS       TRANSMISSION_TYPE = 3
	ES_AIRBORNE_VEL       TRANSMISSION_TYPE = 4
	SURVEILLANCE_ALT      TRANSMISSION_TYPE = 5
	SURVEILLANCE_ID       TRANSMISSION_TYPE = 6
	AIR_TO_AIR            TRANSMISSION_TYPE = 7
	ALL_CALL_REPLY        TRANSMISSION_TYPE = 8

	SELECTION_CHANGE MESSAGE_TYPE = "SEL"
	NEW_ID           MESSAGE_TYPE = "ID"
	NEW_AIRCRAFT     MESSAGE_TYPE = "AIR"
	STATUS_CHANGE    MESSAGE_TYPE = "STA"
	CLICK            MESSAGE_TYPE = "CLK"
	TRANSMISSION     MESSAGE_TYPE = "MSG"
)

type ADSBMessage struct {
	MessageType          MESSAGE_TYPE
	TransmissionType     TRANSMISSION_TYPE
	SessionId            string
	AircraftId           string
	HexIdent             string
	FlightId             string
	DateMessageGenerated string
	TimeMessageGenerated string
	DateMessageLogged    string
	TimeMessageLogged    string
	CallSign             string
	Altitude             int
	GroundSpeed          int
	Track                int
	Latitude             int
	Longitude            int
	VerticalRate         int
	Squawk               string
	Alert                bool
	Emergency            bool
	Spi                  bool
	IsOnGround           bool
}
