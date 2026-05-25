package main

import (
	"github.com/Olyxz16/goview"
)

// StockSeries holds the OHLC-like data for a single ticker.
type StockSeries struct {
	Ticker string    `json:"ticker"`
	Dates  []string  `json:"dates"`
	Values []float64 `json:"values"`
}

// TickerInfo holds human-readable summary stats.
type TickerInfo struct {
	Ticker    string
	Latest    float64
	ChangePct float64
}

// StockVM is the ViewModel for the stock monitor app.
type StockVM struct {
	goview.BaseVM
	StockData  *goview.Observable[StockSeries]
	TickerInfo *goview.Observable[TickerInfo]
}

// NewStockVM creates a ViewModel with empty initial state.
func NewStockVM() *StockVM {
	vm := &StockVM{BaseVM: goview.NewBaseVM()}
	vm.StockData = goview.Observe(StockSeries{}, vm)
	vm.TickerInfo = goview.Observe(TickerInfo{}, vm)
	return vm
}

// Refresh fetches (mock) data for the given ticker and updates observables.
func (vm *StockVM) Refresh(ticker string) {
	if ticker == "" {
		return
	}
	go func() {
		data := generateMockData(ticker)
		vm.StockData.Set(data)

		info := TickerInfo{Ticker: ticker}
		if len(data.Values) > 0 {
			info.Latest = data.Values[len(data.Values)-1]
			if len(data.Values) > 1 {
				prev := data.Values[len(data.Values)-2]
				if prev != 0 {
					info.ChangePct = (info.Latest - prev) / prev * 100
				}
			}
		}
		vm.TickerInfo.Set(info)
	}()
}

// Randomize generates new mock data for the current ticker.
func (vm *StockVM) Randomize() {
	current := vm.StockData.Get()
	if current.Ticker == "" {
		vm.Refresh("DEMO")
		return
	}
	vm.Refresh(current.Ticker)
}
