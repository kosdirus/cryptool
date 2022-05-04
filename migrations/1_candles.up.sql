CREATE TABLE IF NOT EXISTS public.candles
(
    id SERIAL,
    my_id text COLLATE pg_catalog."default" NOT NULL,
    coin_tf text COLLATE pg_catalog."default" NOT NULL,
    coin text COLLATE pg_catalog."default" NOT NULL,
    timeframe text COLLATE pg_catalog."default" NOT NULL,
    utc_open_time timestamp without time zone NOT NULL,
    open_time bigint NOT NULL,
    open double precision NOT NULL,
    high double precision NOT NULL,
    low double precision NOT NULL,
    close double precision NOT NULL,
    volume double precision NOT NULL,
    utc_close_time timestamp without time zone NOT NULL,
    close_time bigint NOT NULL,
    quote_asset_volume double precision NOT NULL,
    number_of_trades bigint NOT NULL,
    taker_buy_base_asset_volume double precision NOT NULL,
    taker_buy_quote_asset_volume double precision NOT NULL,
    ma50 double precision NOT NULL,
    ma50trend boolean NOT NULL,
    ma100 double precision NOT NULL,
    ma100trend boolean NOT NULL,
    ma200 double precision NOT NULL,
    ma200trend boolean NOT NULL,
    CONSTRAINT candles_pkey PRIMARY KEY (my_id)
    )

    TABLESPACE pg_default;