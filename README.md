# CrypTool - Pet project in Golang
App to receive, store, process and send data about crypto assets 
on Binance using API and many other great technologies.

Cryptool is app which is connected to PostgreSQL database and 
through HTTP server (Echo) provides some endpoints with REST API 
(now it's limited connection to some IPs).  
This app was created for collecting data from Binance cryptocurrency
exchange, store them, process and analyze data and then send data 
with API. Such functionality is already working, however there are 
a lot of things in backlog waiting for their time. 

## Purposes:
1. Explore Back-end world and programming languages, especially Golang
2. Experiment and apply in practice acquired theoretical knowledge
3. Create something complete and useful
4. Learn databases both SQL and NoSQL
5. See gaps in knowledge/understanding and work them out in practice

## Used technologies/services
1. Go / Golang
2. Git
3. Goland
4. HTTP server - Echo web framework
5. REST API
6. PostgreSQL (previously was used MongoDB in this project, but to learn 
SQL - database was changed to Postgres)
7. Docker
8. CI/CD (with GitHub Actions and Digital Ocean. Previously Heroku was used)
9. Linux (moved from Windows)

## How does it work
App collects data from Binance through API and store them in Postgres database. 
After request from [Telegram Bot App](https://github.com/kosdirus/tgcrypto), 
Cryptool processing data from database and response with required data.  


## How to test Cryptool
You can go to https://t.me/KynselBot (Telegram) and try one of the following 
commands:
* "sdd 20220531" or "sdd 220531" to receive cryptocurrencies and their price 
change comparing current price and highest price on given date (for example 2022.05.31). 
You can use any date in the last 400 days. "sdd" stands for strong during downtrend - so
  it's better to use if overall now prices are lower than on chosen date.
* "sdu 20220531" or "sdd 220531" to receive cryptocurrencies and their price
  change comparing current price and highest price on given date (for example 2022.05.31).
  You can use any date in the last 400 days. "sdu" stands for strong during uptrend - so 
it's better to use if overall now prices are higher than on chosen date.
* "btcusdt 1d" or "ETHUSDT 30m" to receive candle's data for provided trade pair (coin)
and timeframe. Full list of available trade pairs and timeframes you can find 
[here](https://github.com/kosdirus/cryptool/tree/master/internal/storage/psql/initdata)