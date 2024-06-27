package response

type Login struct {
	Token         string `json:"token"`
	EntityID      uint32 `json:"entity_id"`
	GateAdr       string `json:"gate_adr"`
	GateSocketAdr string `json:"gate_socket_adr"`
}
