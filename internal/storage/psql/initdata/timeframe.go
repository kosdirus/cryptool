package initdata

// Timeframe in this project is the duration of time of a single price bar (candle/candlestick) on a chart.
// Timeframes can be added or deleted in below slice and maps for whole app.
var (
	//TimeframeList = []string{"30m", "1h", "2h", "4h", "8h", "12h", "1d"}

	// TimeframeMap string represents Binance's timeframe and int represents minutes in given timeframe.
	TimeframeMap = map[string]int{
		"30m": 30,
		"1h":  60,
		"2h":  120,
		"4h":  240,
		"8h":  480,
		"12h": 720,
		"1d":  1440,
	}

	// TimeframeMapReverse reverse version of TimeframeMap for better performance in some cases.
	TimeframeMapReverse = map[int]string{
		30:   "30m",
		60:   "1h",
		120:  "2h",
		240:  "4h",
		480:  "8h",
		720:  "12h",
		1440: "1d",
	}

	// TimeframeDays int represents amount of days to download data from Binance in .zip format.
	// Now int for each timeframe is enough to calculate ma50,ma100,ma200.
	TimeframeDays = map[string]int{
		"30m": 5,
		"1h":  9,
		"2h":  18,
		"4h":  35,
		"8h":  69,
		"12h": 400,
		"1d":  400,
	}
)
