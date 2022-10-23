package main

import (
	"time"
	"fmt"
	"sort"
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

type Orders []*Order

func (o Orders) Len() int { return len(o) }
func (o Orders) Swap(i, j int) { o[i], o[j] = o[j], o[i] }
func (o Orders) Less(i, j int) bool { return o[i].Timestamp < o[j].Timestamp } // FIFO order queue

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
	Orders Orders // arr of orders at the target Price | replace []*Order with Order to resolve compiler error not detecting that Orders == []*Order
	TotalVolume float64 // total Size of orders
}

type Limits []*Limit

type ByBestAsk struct{ Limits }

func (a ByBestAsk) Len() int { return len(a.Limits) }
func (a ByBestAsk) Swap(i, j int) { a.Limits[i], a.Limits[j] = a.Limits[j], a.Limits[i] }
func (a ByBestAsk) Less(i, j int) bool { return a.Limits[i].Price < a.Limits[j].Price } // sort asks low to high. counterparty bid matches lowest ask

type ByBestBid struct{ Limits }

func (b ByBestBid) Len() int { return len(b.Limits) }
func (b ByBestBid) Swap(i, j int) { b.Limits[i], b.Limits[j] = b.Limits[j], b.Limits[i] }
func (b ByBestBid) Less(i, j int) bool { return b.Limits[i].Price > b.Limits[j].Price } // sort bids high to low. counterparty ask matches highest bid

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

	// re-sort updated order of Orders after removing o
	sort.Sort(l.Orders)
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