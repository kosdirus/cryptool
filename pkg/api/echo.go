package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/cmd/interal/candle"
	"github.com/kosdirus/cryptool/pkg/binanceapi"
	"github.com/kosdirus/cryptool/pkg/binancezip"
	"github.com/kosdirus/cryptool/pkg/tglogic"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
	"time"
)

func EchoApi(pgdb *pg.DB) {
	// Echo instance
	e := echo.New()

	var BinanceAPIrun uint32
	go binanceapi.BinanceAPISchedule(pgdb, &BinanceAPIrun)

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

	routes := e.Router()
	routes.Add("GET", "/test1", hello)
	// Start server
	port := "80"
	if os.Getenv("ENV") == "LOCAL" {
		port = os.Getenv("IPORT")
	}
	e.Logger.Fatal(e.Start(fmt.Sprint(":", port)))
}

func deleteCandleByMyIDecho(c echo.Context) error {
	candleMyID := c.Param("candleMyID")

	// get the database from context
	pgdb, ok := c.Get("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// delete the candle
	err := candle.DeleteCandle(pgdb, candleMyID)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err = json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  nil,
	}

	_ = json.NewEncoder(c.Response()).Encode(res)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

// Handlers

func hello(c echo.Context) error {
	return c.String(http.StatusOK, fmt.Sprint("Hello, World 👋!\n"))
}

// Returns 1 last closed candles for 30m timeframe for each symbol/coin trade pair
func getCandlesEcho(c echo.Context) error {
	// get database from context
	pgdb, ok := c.Get("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	t := time.Now().UnixMilli()
	tint := t - t%(30*60000) - 30*60000
	candles, err := candle.GetCandleByTimeframeDate(pgdb, "30m", tint)
	if err != nil {
		res := &CandlesResponse{
			Success: false,
			Error:   err.Error(),
			Candles: nil,
		}
		err := json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// send response
	res := &CandlesResponse{
		Success: true,
		Error:   "",
		Candles: candles,
	}

	_ = json.NewEncoder(c.Response()).Encode(res)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

func getCandleByMyIDecho(c echo.Context) error {
	candleMyID := c.Param("candleMyID")

	// get the database from context
	pgdb, ok := c.Get("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(c.Response().Writer).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// query for the candleInstance
	candleInstance, err := candle.GetCandleByMyID(pgdb, candleMyID)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err := json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  candleInstance,
	}

	_ = json.NewEncoder(c.Response()).Encode(res)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

func updateCandleByMyIDecho(c echo.Context) error {
	candleMyID := c.Param("candleMyID")

	// parse in the request body
	req := &UpdateCandleByMyIDRequest{}
	err := json.NewDecoder(c.Request().Body).Decode(req)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err = json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// get the database from context
	pgdb, ok := c.Get("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err = json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// update the candle
	candleInstance, err := candle.UpdateCandle(pgdb, &candle.Candle{
		ID:                       req.ID,
		MyID:                     candleMyID,
		CoinTF:                   req.CoinTF,
		Coin:                     req.Coin,
		Timeframe:                req.Timeframe,
		UTCOpenTime:              req.UTCOpenTime,
		OpenTime:                 req.OpenTime,
		Open:                     req.Open,
		High:                     req.High,
		Low:                      req.Low,
		Close:                    req.Close,
		Volume:                   req.Volume,
		UTCCloseTime:             req.UTCCloseTime,
		CloseTime:                req.CloseTime,
		QuoteAssetVolume:         req.QuoteAssetVolume,
		NumberOfTrades:           req.NumberOfTrades,
		TakerBuyBaseAssetVolume:  req.TakerBuyBaseAssetVolume,
		TakerBuyQuoteAssetVolume: req.TakerBuyQuoteAssetVolume,
		MA50:                     req.MA50,
		MA50Trend:                req.MA50Trend,
		MA100:                    req.MA100,
		MA100Trend:               req.MA100Trend,
		MA200:                    req.MA200,
		MA200Trend:               req.MA200Trend,
	})
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err = json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)

		return err
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  candleInstance,
	}

	_ = json.NewEncoder(c.Response()).Encode(res)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

func createCandleEcho(c echo.Context) error {
	// parse in the request body
	req := &CreateCandleRequest{}
	err := json.NewDecoder(c.Request().Body).Decode(req)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err = json.NewEncoder(c.Response().Writer).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// get the database from context
	pgdb, ok := c.Get("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err = json.NewEncoder(c.Response().Writer).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// insert our candle
	candleinst, err := candle.CreateCandle(pgdb, &candle.Candle{
		MyID:                     req.MyID,
		CoinTF:                   req.CoinTF,
		Coin:                     req.Coin,
		Timeframe:                req.Timeframe,
		UTCOpenTime:              req.UTCOpenTime,
		OpenTime:                 req.OpenTime,
		Open:                     req.Open,
		High:                     req.High,
		Low:                      req.Low,
		Close:                    req.Close,
		Volume:                   req.Volume,
		UTCCloseTime:             req.UTCCloseTime,
		CloseTime:                req.CloseTime,
		QuoteAssetVolume:         req.QuoteAssetVolume,
		NumberOfTrades:           req.NumberOfTrades,
		TakerBuyBaseAssetVolume:  req.TakerBuyBaseAssetVolume,
		TakerBuyQuoteAssetVolume: req.TakerBuyQuoteAssetVolume,
		MA50:                     0.0,
		MA50Trend:                false,
		MA100:                    0.0,
		MA100Trend:               false,
		MA200:                    0.0,
		MA200Trend:               false,
	})
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err = json.NewEncoder(c.Response().Writer).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  candleinst,
	}
	_ = json.NewEncoder(c.Response().Writer).Encode(res)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

func tgGetSymbolEcho(c echo.Context) error {
	symbol := c.Param("symbol")
	timeframe := c.Param("timeframe")

	pgdb := c.Get("DB").(*pg.DB)

	// query for the candleInstance
	candleInstance, err := candle.GetCandleBySymbolAndTimeframe(pgdb, symbol, timeframe)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err := json.NewEncoder(c.Response()).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		c.Response().WriteHeader(http.StatusBadRequest)
		return err
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  candleInstance,
	}

	_ = json.NewEncoder(c.Response()).Encode(res)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

func sddEcho(c echo.Context) error {
	userTime := c.Param("time")
	// get the database from context
	pgdb := c.Get("DB").(*pg.DB)

	// query for the candleInstance
	candleMap := tglogic.StrongDuringDowntrend(pgdb, userTime)

	_ = json.NewEncoder(c.Response()).Encode(candleMap)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

func sduEcho(c echo.Context) error {
	userTime := c.Param("time")
	// get the database from context
	pgdb := c.Get("DB").(*pg.DB)

	// query for the candleInstance
	candleMap := tglogic.StrongDuringUptrend(pgdb, userTime)

	_ = json.NewEncoder(c.Response()).Encode(candleMap)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

func binanceAPIhandler(c echo.Context) error {
	pgdb := c.Get("DB").(*pg.DB)
	BinanceAPIrun := c.Get("BinanceAPIrun").(*uint32)
	go func() {
		binanceapi.BinanceAPI(pgdb, BinanceAPIrun)
	}()

	return c.String(http.StatusOK, "BinanceAPIToPostgres migration started! Have some rest!")
}

func binanceZipHandler(c echo.Context) error {
	pgdb := c.Get("DB").(*pg.DB)

	go func() {
		binancezip.BinanceZipToPostgres(pgdb)
	}()

	return c.String(http.StatusOK, "BinanceZipToPostgres migration started! Have some rest!\n")
}

func dbIntegrityHandler(c echo.Context) error {
	pgdb := c.Get("DB").(*pg.DB)

	go func() {
		binanceapi.CheckDBIntegrity(pgdb)
	}()

	return c.String(http.StatusOK, "Database integrity is being checked!\n")
}
