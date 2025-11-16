package constant

type OrderStatus int

const (
	OrderStatusPending   OrderStatus = 1
	OrderStatusCompleted OrderStatus = 2
	OrderStatusCanceled  OrderStatus = 3
)
