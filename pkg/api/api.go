package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-pg/pg/v10"
	"github.com/kosdirus/cryptool/cmd/interal/candle"
	"github.com/kosdirus/cryptool/pkg/binanceapi"
	"github.com/kosdirus/cryptool/pkg/binancezip"
	"github.com/kosdirus/cryptool/pkg/tglogic"
	"log"
	"net/http"
	"time"
)

func NewAPI(pgdb *pg.DB) *chi.Mux {
	// setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.WithValue("DB", pgdb))

	var BinanceAPIrun uint32

	go binanceapi.BinanceAPISchedule(pgdb, &BinanceAPIrun)

	r.Route("/candles", func(r chi.Router) {
		r.Post("/", createCandle)
		r.Get("/{candleMyID}", getCandleByMyID)
		r.Get("/", getCandles)
		r.Put("/{candleMyID}", updateCandleByMyID)
		r.Delete("/{candleMyID}", deleteCandleByMyID)
	})

	r.Route("/tg", func(r chi.Router) {
		r.Get("/{symbol}/{timeframe}", tgGetSymbol)
		r.Get("/sdd/{time}", sdd)
		r.Get("/sdu/{time}", sdu)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World 👋!"))
	})

	r.Get("/dbintegrity", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Database integrity is being checked"))
		go binanceapi.CheckDBIntegrity(pgdb)
		time.Sleep(50 * time.Millisecond)
	})

	r.Get("/binancezip", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("BinanceZipToPostgres migration started! Have some rest!"))
		go binancezip.BinanceZipToPostgres(pgdb)
		time.Sleep(400 * time.Millisecond)
	})

	r.Get("/binanceapi", func(w http.ResponseWriter, r *http.Request) {
		go func() {
			binanceapi.BinanceAPI(pgdb, &BinanceAPIrun)
		}()
		w.Write([]byte("BinanceAPIToPostgres migration started! Have some rest!"))
		time.Sleep(400 * time.Millisecond)
	})

	return r
}

func tgGetSymbol(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	timeframe := chi.URLParam(r, "timeframe")

	// get the database from context
	pgdb, ok := r.Context().Value("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// query for the candleInstance
	candleInstance, err := candle.GetCandleBySymbolAndTimeframe(pgdb, symbol, timeframe)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  candleInstance,
	}

	_ = json.NewEncoder(w).Encode(res)
	w.WriteHeader(http.StatusOK)
}

func sdd(w http.ResponseWriter, r *http.Request) {
	userTime := chi.URLParam(r, "time")
	// get the database from context
	pgdb, ok := r.Context().Value("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// query for the candleInstance
	candleMap := tglogic.StrongDuringDowntrend(pgdb, userTime)

	_ = json.NewEncoder(w).Encode(candleMap)
	w.WriteHeader(http.StatusOK)
}

func sdu(w http.ResponseWriter, r *http.Request) {
	userTime := chi.URLParam(r, "time")
	// get the database from context
	pgdb, ok := r.Context().Value("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// query for the candleInstance
	candleMap := tglogic.StrongDuringUptrend(pgdb, userTime)

	_ = json.NewEncoder(w).Encode(candleMap)
	w.WriteHeader(http.StatusOK)
}

type CreateCandleRequest struct {
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
	Success bool           `json:"success"`
	Error   string         `json:"error"`
	Candle  *candle.Candle `json:"candle"`
}

type CreateCandleResponse struct {
}

func createCandle(w http.ResponseWriter, r *http.Request) {
	// parse in the request body
	req := &CreateCandleRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get the database from context
	pgdb, ok := r.Context().Value("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
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
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  candleinst,
	}

	_ = json.NewEncoder(w).Encode(res)
	w.WriteHeader(http.StatusOK)
}

func getCandleByMyID(w http.ResponseWriter, r *http.Request) {
	candleMyID := chi.URLParam(r, "candleMyID")

	// get the database from context
	pgdb, ok := r.Context().Value("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// query for the candleInstance
	candleInstance, err := candle.GetCandleByMyID(pgdb, candleMyID)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  candleInstance,
	}

	_ = json.NewEncoder(w).Encode(res)
	w.WriteHeader(http.StatusOK)
}

type CandlesResponse struct {
	Success bool            `json:"success"`
	Error   string          `json:"error"`
	Candles []candle.Candle `json:"candles"`
}

func getCandles(w http.ResponseWriter, r *http.Request) {
	// get database from context
	pgdb, ok := r.Context().Value("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	candles, err := candle.GetCandles(pgdb)
	if err != nil {
		res := &CandlesResponse{
			Success: false,
			Error:   err.Error(),
			Candles: nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// send response
	res := &CandlesResponse{
		Success: true,
		Error:   "",
		Candles: candles,
	}

	_ = json.NewEncoder(w).Encode(res)
	w.WriteHeader(http.StatusOK)
}

type UpdateCandleByMyIDRequest struct {
	ID                       int64     `json:"id"`
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

func updateCandleByMyID(w http.ResponseWriter, r *http.Request) {
	candleMyID := chi.URLParam(r, "candleMyID")

	// parse in the request body
	req := &UpdateCandleByMyIDRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get the database from context
	pgdb, ok := r.Context().Value("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
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
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  candleInstance,
	}

	_ = json.NewEncoder(w).Encode(res)
	w.WriteHeader(http.StatusOK)

}

func deleteCandleByMyID(w http.ResponseWriter, r *http.Request) {
	candleMyID := chi.URLParam(r, "candleMyID")

	// get the database from context
	pgdb, ok := r.Context().Value("DB").(*pg.DB)
	if !ok {
		res := &CandleResponse{
			Success: false,
			Error:   "could not get database from context",
			Candle:  nil,
		}
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// delete the candle
	err := candle.DeleteCandle(pgdb, candleMyID)
	if err != nil {
		res := &CandleResponse{
			Success: false,
			Error:   err.Error(),
			Candle:  nil,
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Printf("error sending response: %v\n", err)
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// return a response
	res := &CandleResponse{
		Success: true,
		Error:   "",
		Candle:  nil,
	}

	_ = json.NewEncoder(w).Encode(res)
	w.WriteHeader(http.StatusOK)
}
