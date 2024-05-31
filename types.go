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
	MessageType          MESSAGE_TYPE      `json:"message_type"`
	TransmissionType     TRANSMISSION_TYPE `json:"transmission_type"`
	SessionId            string            `json:"session_id"`
	AircraftId           string            `json:"aircraft_id"`
	HexIdent             string            `json:"hex_ident"`
	FlightId             string            `json:"flight_id"`
	DateMessageGenerated string            `json:"date_message_generated"`
	TimeMessageGenerated string            `json:"time_message_generated"`
	DateMessageLogged    string            `json:"date_message_logged"`
	TimeMessageLogged    string            `json:"time_message_logged"`
	CallSign             string            `json:"call_sign"`
	Altitude             int               `json:"altitude"`
	GroundSpeed          int               `json:"ground_speed"`
	Track                int               `json:"track"`
	Latitude             int               `json:"latitude"`
	Longitude            int               `json:"longitude"`
	VerticalRate         int               `json:"vertical_rate"`
	Squawk               string            `json:"squawk"`
	Alert                bool              `json:"alert"`
	Emergency            bool              `json:"emergency"`
	Spi                  bool              `json:"spi"`
	IsOnGround           bool              `json:"is_on_ground"`
}
