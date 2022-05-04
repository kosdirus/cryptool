package symbol

var (
	SymbolList = []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "XRPUSDT", "LUNAUSDT", "ADAUSDT", "SOLUSDT", "EGLDUSDT",
		"SRMUSDT", "TWTUSDT", "AVAXUSDT", "DOTUSDT", "DOGEUSDT", "SHIBUSDT", "MATICUSDT", "ATOMUSDT", "LTCUSDT",
		"NEARUSDT", "LINKUSDT", "TRXUSDT", "UNIUSDT", "FTTUSDT", "BCHUSDT", "ALGOUSDT", "MANAUSDT", "XLMUSDT",
		"HBARUSDT", "FTMUSDT", "ETCUSDT", "ICPUSDT", "SANDUSDT", "FILUSDT", "VETUSDT", "AXSUSDT", "KLAYUSDT",
		"XMRUSDT", "THETAUSDT", "XTZUSDT", "HNTUSDT", "WAVESUSDT", "EOSUSDT", "IOTAUSDT", "FLOWUSDT", "BTTCUSDT",
		"AAVEUSDT", "CAKEUSDT", "GRTUSDT", "ONEUSDT", "GALAUSDT", "ZECUSDT", "RUNEUSDT", "NEOUSDT", "STXUSDT",
		"QNTUSDT", "XECUSDT", "CHZUSDT", "CELOUSDT", "ENJUSDT", "ANCUSDT", "AMPUSDT", "KSMUSDT", "ARUSDT",
		"BATUSDT", "CRVUSDT", "LRCUSDT", "DASHUSDT", "XEMUSDT", "CVXUSDT", "ROSEUSDT", "TFUELUSDT", "MINAUSDT",
		"DCRUSDT", "SCRTUSDT", "HOTUSDT", "YFIUSDT", "COMPUSDT", "IOTXUSDT", "SXPUSDT", "QTUMUSDT", "BNTUSDT",
		"ANKRUSDT", "RVNUSDT", "UMAUSDT", "RNDRUSDT", "1INCHUSDT", "PAXGUSDT", "OMGUSDT", "BTGUSDT", "GLMRUSDT",
		"KAVAUSDT", "ZILUSDT", "LPTUSDT", "ICXUSDT", "ONTUSDT",
		"AUDIOUSDT", "WOOUSDT", "SCUSDT", "KNCUSDT", "SNXUSDT", "ZENUSDT", "ZRXUSDT", "IOSTUSDT", "SKLUSDT",
		"SUSHIUSDT", "STORJUSDT", "RENUSDT", "POLYUSDT", "IMXUSDT", "HIVEUSDT", "SYSUSDT", "JSTUSDT", "ILVUSDT",
		"CKBUSDT", "SPELLUSDT", "DYDXUSDT", "FXSUSDT", "FLUXUSDT", "PERPUSDT", "PEOPLEUSDT", "ENSUSDT", "DGBUSDT",
		"OCEANUSDT", "WINUSDT", "INJUSDT", "SUPERUSDT", "FETUSDT", "TRIBEUSDT", "CELRUSDT", "LSKUSDT", "PLAUSDT",
		"POWRUSDT" /*"ANYUSDT",*/, "PYRUSDT", "YGGUSDT", "C98USDT", "WRXUSDT", "DENTUSDT", "XNOUSDT", "RAYUSDT",
		"COTIUSDT", "API3USDT", "CHRUSDT", "MBLUSDT", "ALICEUSDT", "REQUSDT", "ONGUSDT", "BTCSTUSDT", "JOEUSDT",
		"MDXUSDT", "PUNDIXUSDT", "ANTUSDT", "ARDRUSDT", "MOVRUSDT", "ASTRUSDT", "ACHUSDT", "CFXUSDT",
		"CVCUSDT", "CTSIUSDT", "REEFUSDT", "MBOXUSDT", "RSRUSDT", "BETAUSDT", "ORNUSDT", "ALPHAUSDT", "BANDUSDT",
		"NKNUSDT", "DUSKUSDT", "SUNUSDT", "RIFUSDT", "FUNUSDT", "MASKUSDT", "POLSUSDT", "TOMOUSDT", "BAKEUSDT"}
	Timeframe    = []string{"30m", "1h", "2h", "4h", "8h", "12h", "1d"}
	TimeframeMap = map[int]string{
		30:   "30m",
		60:   "1h",
		120:  "2h",
		240:  "4h",
		480:  "8h",
		720:  "12h",
		1440: "1d",
	}
	/*TimeframeDays = map[string]int{
		"30m": 2,
		"1h":  3,
		"2h":  4,
		"4h":  5,
		"8h":  6,
		"12h": 10,
		"1d":  10,
	}*/
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
