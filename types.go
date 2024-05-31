package main

type ADSBMessage struct {
	MessageType          string `json:"message_type"`
	TransmissionType     int    `json:"transmission_type"`
	SessionId            string `json:"session_id"`
	AircraftId           string `json:"aircraft_id"`
	HexIdent             string `json:"hex_ident"`
	FlightId             string `json:"flight_id"`
	DateMessageGenerated string `json:"date_message_generated"`
	TimeMessageGenerated string `json:"time_message_generated"`
	DateMessageLogged    string `json:"date_message_logged"`
	TimeMessageLogged    string `json:"time_message_logged"`
	CallSign             string `json:"call_sign"`
	Altitude             int    `json:"altitude"`
	GroundSpeed          int    `json:"ground_speed"`
	Track                int    `json:"track"`
	Latitude             int    `json:"latitude"`
	Longitude            int    `json:"longitude"`
	VerticalRate         int    `json:"vertical_rate"`
	Squawk               string `json:"squawk"`
	Alert                bool   `json:"alert"`
	Emergency            bool   `json:"emergency"`
	Spi                  bool   `json:"spi"`
	IsOnGround           bool   `json:"is_on_ground"`
}
