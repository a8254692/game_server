package response

type Order struct {
	EntityID uint32 `json:"entity_id"`
	OrderId  string `json:"order_id"`
}
