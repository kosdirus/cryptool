package rest

import (
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/core"
	"github.com/kosdirus/cryptool/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"os"
	"time"
)

// CandleRequest probably default "Candle" struct is the same, so it can be enough to use it here
type CandleRequest struct {
	ID                       int64     `json:"id"`
	MyID                     string    `bson:"my_id" json:"my_id" pg:"my_id"`
	CoinTF                   string    `bson:"coin_tf" json:"coin_tf" pg:"coin_tf"`
	Coin                     string    `bson:"coin" json:"coin" pg:"coin"`
	Timeframe                string    `bson:"timeframe" json:"timeframe" pg:"timeframe"`
	UTCOpenTime              time.Time `bson:"utc_open_time" json:"utc_open_time" pg:"utc_open_time"`
	OpenTime                 int64     `bson:"open_time" json:"open_time" pg:"open_time"`
	Open                     float64   `bson:"open" json:"open" pg:"open"`
	High                     float64   `bson:"high" json:"high" pg:"high"`
	Low                      float64   `bson:"low" json:"low" pg:"low"`
	Close                    float64   `bson:"close" json:"close" pg:"close"`
	Volume                   float64   `bson:"volume" json:"volume" pg:"volume"`
	UTCCloseTime             time.Time `bson:"utc_close_time" json:"utc_close_time" pg:"utc_close_time"`
	CloseTime                int64     `bson:"close_time" json:"close_time" pg:"close_time"`
	QuoteAssetVolume         float64   `bson:"quote_asset_volume" json:"quote_asset_volume" pg:"quote_asset_volume"`
	NumberOfTrades           int64     `bson:"number_of_trades" json:"number_of_trades" pg:"number_of_trades"`
	TakerBuyBaseAssetVolume  float64   `bson:"taker_buy_base_asset_volume" json:"taker_buy_base_asset_volume" pg:"taker_buy_base_asset_volume"`
	TakerBuyQuoteAssetVolume float64   `bson:"taker_buy_quote_asset_volume" json:"taker_buy_quote_asset_volume" pg:"taker_buy_quote_asset_volume"`
	MA50                     float64   `bson:"ma50" json:"ma50" pg:"ma50,use_zero"`
	MA50Trend                bool      `bson:"ma50trend" json:"ma50trend" pg:"ma50trend,use_zero"`
	MA100                    float64   `bson:"ma100" json:"ma100" pg:"ma100,use_zero"`
	MA100Trend               bool      `bson:"ma100trend" json:"ma100trend" pg:"ma100trend,use_zero"`
	MA200                    float64   `bson:"ma200" json:"ma200" pg:"ma200,use_zero"`
	MA200Trend               bool      `bson:"ma200trend" json:"ma200trend" pg:"ma200trend,use_zero"`
}

type CandleResponse struct {
	Success bool         `json:"success"`
	Error   string       `json:"error"`
	Candle  *core.Candle `json:"candle"`
}

type CandlesResponse struct {
	Success bool          `json:"success"`
	Error   string        `json:"error"`
	Candles []core.Candle `json:"candles"`
}

// EchoApiServer receives pointer to go-pg database pool of connection.
// It uses middleware, registers routers and starts an HTTP server.
func EchoApiServer(pgdb *pg.DB) {
	// Setup echo router instance
	e := echo.New()

	// Schedule continuous data acquisition from Binance
	var BinanceAPIrun uint32
	go service.BinanceAPISchedule(pgdb, &BinanceAPIrun)

	// Middleware
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("DB", pgdb)
			c.Set("BinanceAPIrun", &BinanceAPIrun)
			return next(c)
		}
	})
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//Block requests received not from list of allowed IP addresses
	//If you want to give access to all IP addresses - just comment or delete "if block" below.
	if os.Getenv("ENV") == "DIGITAL" {
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				//Map for allowed IP addresses is used here for convenience
				//for those who needs many IP addresses to be allowed
				allowedIPaddresses := map[string]struct{}{
					os.Getenv("LAPTOPIP"): {},
					os.Getenv("TGBOTIP"):  {},
				}
				if _, exists := allowedIPaddresses[c.RealIP()]; !exists {
					return echo.NewHTTPError(http.StatusUnauthorized,
						fmt.Sprintf("IP address %s not allowed", c.RealIP()))
				}
				return next(c)
			}
		})
	}

	// Routes
	e.GET("/", hello)
	e.GET("/binanceapi", binanceAPIhandler)
	e.GET("/binancezip", binanceZipHandler)
	e.GET("/dbintegrity", dbIntegrityHandler)

	g := e.Group("/candles")
	g.POST("/", createCandleEcho)
	g.GET("/:candleMyID", getCandleByMyIDecho)
	g.GET("/", getCandlesEcho)
	g.PUT("/:canldeMyID", updateCandleByMyIDecho)
	g.DELETE("/:candleMyID", deleteCandleByMyIDecho)

	g = e.Group("/tg")
	g.GET("/:symbol/:timeframe", tgGetSymbolEcho)
	g.GET("/sdd/:time", sddEcho)
	g.GET("/sdu/:time", sduEcho)

	// Start server
	port := "80"
	if os.Getenv("ENV") == "LOCAL" {
		port = os.Getenv("IPORT")
	}
	e.Logger.Fatal(e.Start(fmt.Sprint(":", port)))
}

// Handlers
