package main

import (
	"math/rand"
	"time"

	"github.com/Olyxz16/goview"
	"github.com/Olyxz16/goview-echart/driver"
)

func main() {
	vm := NewStockVM()

	d := driver.New()
	defer d.Destroy()
	d.SetTitle("Stock monitor")
	d.SetSize(900, 600)

	eval := vm.Eval(d.Eval)
	app := goview.NewApp(d, eval, vm.Dispatch())

	app.Mount(
		goview.NewPatchComponent(
			"#stock-chart",
			vm.StockData,
			renderStockChart,
			eval,
		),
		goview.NewComponent(
			"#ticker-info",
			vm.TickerInfo,
			renderTickerInfo,
			eval,
		),
	)

	app.Bind("Refresh", vm.BindString(vm.Refresh))
	app.Bind("Randomize", vm.BindVoid(vm.Randomize))

	// Load initial demo data so the chart renders immediately.
	vm.Refresh("DEMO")

	app.Run("index.html")
}

// generateMockData creates random-walk stock data for demo purposes.
func generateMockData(ticker string) StockSeries {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	const days = 30
	base := 100.0 + rng.Float64()*50.0

	dates := make([]string, days)
	values := make([]float64, days)

	now := time.Now()
	for i := 0; i < days; i++ {
		d := now.AddDate(0, 0, -days+i+1)
		dates[i] = d.Format("2006-01-02")
		change := (rng.Float64() - 0.5) * 5.0
		base += change
		if base < 10 {
			base = 10
		}
		values[i] = base
	}

	return StockSeries{
		Ticker: ticker,
		Dates:  dates,
		Values: values,
	}
}
