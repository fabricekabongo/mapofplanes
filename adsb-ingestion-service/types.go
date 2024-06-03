package main

const (
	MessageTypeSelectionChange = "SEL"
	MessageTypeNewID           = "ID"
	MessageTypeNewAircraft     = "AIR"
	MessageTypeStatusChange    = "STA"
	MessageTypeClick           = "CLK"
	MessageTypeTransmission    = "MSG"
)

const (
	TransmissionTypeIdentityAndCategory = 1
	TranmissionTypeSurfacePosition      = 2
	TranmissionTypeAirbornePosition     = 3
	TranmissionTypeAirborneVelocity     = 4
	TranmissionTypeSurveillanceAltitude = 5
	TranmissionTypeSurveillanceId       = 6
	TranmissionTypeAirToAir             = 7
	TranmissionTypeAllCallReply         = 8
)

type ADSBMessage struct {
	MessageType          string  `json:"message_type"`
	TransmissionType     int     `json:"transmission_type"`
	SessionId            string  `json:"session_id"`
	AircraftId           string  `json:"aircraft_id"`
	HexIdent             string  `json:"hex_ident"`
	FlightId             string  `json:"flight_id"`
	DateMessageGenerated string  `json:"date_message_generated"`
	TimeMessageGenerated string  `json:"time_message_generated"`
	DateMessageLogged    string  `json:"date_message_logged"`
	TimeMessageLogged    string  `json:"time_message_logged"`
	CallSign             string  `json:"call_sign"`
	Altitude             float64 `json:"altitude"`
	GroundSpeed          float64 `json:"ground_speed"`
	Track                int     `json:"track"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	VerticalRate         float64 `json:"vertical_rate"`
	Squawk               string  `json:"squawk"`
	Alert                bool    `json:"alert"`
	Emergency            bool    `json:"emergency"`
	Spi                  bool    `json:"spi"`
	IsOnGround           bool    `json:"is_on_ground"`
}
