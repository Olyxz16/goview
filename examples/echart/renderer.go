package main

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/Olyxz16/goview"
)

// renderStockChart returns an AttrPatch that sets data-option on the chart
// container. A MutationObserver in index.html watches this attribute and
// calls chart.setOption() whenever it changes.
func renderStockChart(s StockSeries) goview.Patch {
	option := map[string]any{
		"title": map[string]any{
			"text": s.Ticker + " Price",
		},
		"tooltip": map[string]any{
			"trigger": "axis",
		},
		"xAxis": map[string]any{
			"type": "category",
			"data": s.Dates,
		},
		"yAxis": map[string]any{
			"type":  "value",
			"scale": true,
		},
		"series": []map[string]any{
			{
				"type":      "line",
				"data":      s.Values,
				"smooth":    true,
				"symbol":    "none",
				"lineStyle": map[string]any{"width": 2},
				"areaStyle": map[string]any{
					"opacity": 0.2,
				},
			},
		},
		"grid": map[string]any{
			"left":         "3%",
			"right":        "4%",
			"bottom":       "3%",
			"containLabel": true,
		},
	}
	data, _ := json.Marshal(option)
	return goview.Attr("data-option", string(data))
}

// renderTickerInfo renders a simple HTML summary of the current ticker.
func renderTickerInfo(info TickerInfo) string {
	if info.Ticker == "" {
		return `<div class="empty">Enter a ticker and press Refresh</div>`
	}
	latest := strconv.FormatFloat(info.Latest, 'f', 2, 64)
	change := strconv.FormatFloat(math.Round(info.ChangePct*100)/100, 'f', 2, 64)
	cls := "change"
	if info.ChangePct > 0 {
		cls = "change up"
	} else if info.ChangePct < 0 {
		cls = "change down"
	}
	return `<div class="ticker-card">` +
		`<span class="ticker-name">` + info.Ticker + `</span>` +
		`<span class="latest">` + latest + `</span>` +
		`<span class="` + cls + `">` + change + `%</span>` +
		`</div>`
}
