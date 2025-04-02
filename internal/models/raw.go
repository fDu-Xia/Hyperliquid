package models

// Trade node_trades
type Trade struct {
	Coin            string     `json:"coin"`
	Side            string     `json:"side"`
	Time            string     `json:"time"`
	Price           string     `json:"px"`
	Size            string     `json:"sz"`
	Hash            string     `json:"hash"`
	TradeDirections string     `json:"trade_dir_override"`
	SideInfo        []SideInfo `json:"side_info"`
}

type SideInfo struct {
	User      string  `json:"user"`
	StartPos  string  `json:"start_pos"`
	OrderID   int64   `json:"oid"`
	TwapID    *string `json:"twap_id"`
	ClOrderID *string `json:"cloid"`
}

// OrderStatus node_order_statuses
type OrderStatus struct {
	Time   string `json:"time"`
	User   string `json:"user"`
	Status string `json:"status"`
	Order  Order  `json:"order"`
}

type Order struct {
	Coin             string   `json:"coin"`
	Side             string   `json:"side"`
	LimitPrice       string   `json:"limitPx"`
	Size             string   `json:"sz"`
	OrderID          int64    `json:"oid"`
	Timestamp        int64    `json:"timestamp"`
	TriggerCondition string   `json:"triggerCondition"`
	IsTrigger        bool     `json:"isTrigger"`
	TriggerPrice     string   `json:"triggerPx"`
	Children         []string `json:"children"`
	IsPositionTpsl   bool     `json:"isPositionTpsl"`
	ReduceOnly       bool     `json:"reduceOnly"`
	OrderType        string   `json:"orderType"`
	OrigSize         string   `json:"origSz"`
	TimeInForce      string   `json:"tif"`
	ClOrderID        *string  `json:"cloid"`
}
