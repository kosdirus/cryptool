package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/internal/core"
	"github.com/kosdirus/cryptool/internal/service"
	"github.com/kosdirus/cryptool/internal/storage/psql"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"time"
)

// Handler deleteCandleByMyIDecho receives core.Candle.MyID and delete it from database.
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
	err := psql.DeleteCandle(pgdb, candleMyID)
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

func hello(c echo.Context) error {
	return c.String(http.StatusOK, fmt.Sprint("Hello, World 👋!\n"))
}

// Handler getCandlesEcho returns 1 last closed candles for 30m timeframe for each trade pair.
func getCandlesEcho(c echo.Context) error {
	// get database from context
	pgdb, ok := c.Get("DB").(*pg.DB)
	if !ok {
		res := &CandlesResponse{
			Success: false,
			Error:   "could not get database from context",
			Candles: nil,
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
	candles, err := psql.GetCandleByTimeframeDate(pgdb, "30m", tint)
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

// Handler getCandleByMyIDecho returns core.Candle by MyID given by param.
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
	candleInstance, err := psql.GetCandleByMyID(pgdb, candleMyID)
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

// Handler updateCandleByMyIDecho updates core.Candle by MyID given by param.
func updateCandleByMyIDecho(c echo.Context) error {
	candleMyID := c.Param("candleMyID")

	// parse in the request body
	req := &core.Candle{}
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
	if candleMyID != req.MyID {
		return errors.New("given candle has different MyID than param in URL")
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

	_, err = psql.UpdateCandle(pgdb, req)
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
		Candle:  req,
	}

	_ = json.NewEncoder(c.Response()).Encode(res)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

// Handler createCandleEcho parses request bode and creates core.Candle in database.
func createCandleEcho(c echo.Context) error {
	// parse in the request body
	req := &core.Candle{}
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
	_, err = psql.CreateCandleCheckForExists(pgdb, req)
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
		Candle:  req,
	}
	_ = json.NewEncoder(c.Response().Writer).Encode(res)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

// Handler tgGetSymbolEcho used for telegram bot requests and returns candle for given coin(symbol) and timeframe.
func tgGetSymbolEcho(c echo.Context) error {
	symbol := c.Param("symbol")
	timeframe := c.Param("timeframe")

	pgdb := c.Get("DB").(*pg.DB)

	// query for the candleInstance
	candleInstance, err := psql.GetCandleBySymbolAndTimeframe(pgdb, symbol, timeframe)
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

// Handler sddEcho receive time and pass it to service.StrongDuringDowntrend func.
func sddEcho(c echo.Context) error {
	userTime := c.Param("time")
	// get the database from context
	pgdb := c.Get("DB").(*pg.DB)

	// query for the candleInstance
	candleMap := service.StrongDuringDowntrend(pgdb, userTime)

	_ = json.NewEncoder(c.Response()).Encode(candleMap)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

// Handler sduEcho receive time and pass it to service.StrongDuringUptrend func.
func sduEcho(c echo.Context) error {
	userTime := c.Param("time")
	// get the database from context
	pgdb := c.Get("DB").(*pg.DB)

	// query for the candleInstance
	candleMap := service.StrongDuringUptrend(pgdb, userTime)

	_ = json.NewEncoder(c.Response()).Encode(candleMap)
	c.Response().WriteHeader(http.StatusOK)

	return nil
}

// Handler binanceAPIhandler used to start fetching candlestick data from Binance through API.
func binanceAPIhandler(c echo.Context) error {
	pgdb := c.Get("DB").(*pg.DB)
	BinanceAPIrun := c.Get("BinanceAPIrun").(*uint32)
	go func() {
		service.BinanceAPI(pgdb, BinanceAPIrun)
	}()

	return c.String(http.StatusOK, "BinanceAPIToPostgres migration started! Have some rest!")
}

// Handler binanceZipHandler used to start fetching candlestick data from Binance through zip and CSV.
func binanceZipHandler(c echo.Context) error {
	pgdb := c.Get("DB").(*pg.DB)

	go func() {
		service.BinanceZipToPostgres(pgdb)
	}()

	return c.String(http.StatusOK, "BinanceZipToPostgres migration started! Have some rest!\n")
}

// Handler dbIntegrityHandler used to start checking database for missing candles.
func dbIntegrityHandler(c echo.Context) error {
	pgdb := c.Get("DB").(*pg.DB)

	go func() {
		service.CheckDBIntegrity(pgdb)
	}()

	return c.String(http.StatusOK, "Database integrity is being checked!\n")
}
