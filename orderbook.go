package main

import (
	"time"
	"fmt"
)

// Bid-Ask Order pair
type Match struct {
	Ask *Order
	Bid *Order
	SizeFilled float64 // partial fill > 0 or complete = 0
	Price float64
}

type Order struct {
	Size float64 // qty of asset
	Bid bool // whether order is bid or ask
	Limit *Limit // the limit order that this order belongs to
	Timestamp int64
}

func NewOrder(bid bool, size float64) *Order {
	return &Order{
		Size: size,
		Bid: bid,
		Timestamp: time.Now().UnixNano(),
	}
}

// pretty print Order in console to size 2 decimals
func (o *Order) String() string {
	return fmt.Sprintf("[size: %.2f]", o.Size)
}

// Limit order consists of orders at Price of varied Size
type Limit struct {
	Price float64
	Orders []*Order // arr of orders at the target Price
	TotalVolume float64 // total Size of orders
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price: price,
		Orders: []*Order{},
	}
}

// // string repr of Limit for testing
// func (l *Limit) String() string {
// 	return fmt.Sprintf("[price: %.2f | volume: %.2f]", l.Price, l.TotalVolume)
// }

// add Order to Limit
func (l *Limit) AddOrder(o *Order) {
	o.Limit = l // ref to Limit useful for matching/filling orders
	l.Orders = append(l.Orders, o)
	l.TotalVolume += o.Size
}

// cancel/delete Order from Limit
func (l *Limit) DeleteOrder(o *Order) {
	for i := 0; i < len(l.Orders); i++ {
		if (l.Orders[i] == o) {
			l.Orders[i] = l.Orders[len(l.Orders) - 1] // replace order at i with last order in slice
			l.Orders = l.Orders[:len(l.Orders) - 1] // shift whole arr left by 1: slice from [0:last - 1] to remove duplicate last order
		}
	}
	o.Limit = nil // remove ref
	l.TotalVolume -= o.Size

	// TODO re-sort  orders
}

// collection of orders
type Orderbook struct {
	Asks []*Limit // arr of sell orders
	Bids []*Limit // arr of buy orders

	// map Limit.Price -> Limit for tracking if a Limit already exists at a given Price
	// every price point can have a Limit of 1 or more orders
	AskLimits map[float64]*Limit
	BidLimits map[float64]*Limit
}

func NewOrderbook() *Orderbook {
	return &Orderbook{
		Asks: []*Limit{},
		Bids: []*Limit{},
		AskLimits: make(map[float64]*Limit),
		BidLimits: make(map[float64]*Limit),
	}
}

// match Order
// add Order to Orderbook can return 0 or more order matches
func (ob *Orderbook) PlaceOrder(price float64, o *Order) []Match {
	// 1. try to match the orders
	// matching logic

	// 2. add the rest/remaining partially filled Order to the orderbook
	if o.Size > 0.0 {
		ob.add(price, o)
	}
	return []Match{}
}

// add Order to orderbook
func (ob *Orderbook) add(price float64, o *Order) {
	var limit *Limit

	if o.Bid {
		limit = ob.BidLimits[price] // get Bid Limit Container at Price
	
	} else {
		limit = ob.AskLimits[price] // get Ask Limit Container at Price
	}

	if limit != nil {
		limit.AddOrder(o)
	}

	// if no Limit exists at price point, create a new one
	if limit == nil {
		limit = NewLimit(price)
		limit.AddOrder(o) // add order to Limit orders at price
		// add Limit to orderbook and to orderbook at price point
		if o.Bid {
			ob.Bids = append(ob.Bids, limit)
			ob.BidLimits[price] = limit // add price-limit mapping at price point
		} else {
			ob.Asks = append(ob.Asks, limit)
			ob.AskLimits[price] = limit // add price-limit mapping at price point
		}
	}
}