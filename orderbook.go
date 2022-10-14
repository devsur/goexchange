package main

import (
	"time"
	"fmt"
)

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
	Orders []*Order // arr of partial orders that fill the limit order
	TotalVolume float64 // total Size of orders
}

func NewLimit(price float64) *Limit {
	return &Limit{
		Price: price,
		Orders: []*Order{},
	}
}

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
			l.Orders[i] = l.Orders[len(l.Orders) - 1] // shift current left
			l.Orders = l.Orders[:len(l.Orders) - 1] // shift whole arr left
		}
	}
	o.Limit = nil // remove ref
	l.TotalVolume -= o.Size

	// TODO resort 
}

// collection of orders
type Orderbook struct {
	Asks []*Limit // arr of sell orders
	Bids []*Limit // arr of buy orders
}