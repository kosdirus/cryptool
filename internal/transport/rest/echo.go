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
)

// CandleResponse used for response to HTTP requests for single instance of core.Candle.
type CandleResponse struct {
	Success bool         `json:"success"`
	Error   string       `json:"error"`
	Candle  *core.Candle `json:"candle"`
}

// CandlesResponse used for response to HTTP requests for multiple instances of core.Candle.
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
