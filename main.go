package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
    "github.com/fatih/color"
    "github.com/bitfinexcom/bitfinex-api-go/v1"
)

type MyData struct {
    name string
    amount float64
    priceUsd float64
    color *color.Color
}

func (data *MyData) Print() {
    data.color.Printf("%s: %.4f", data.name, data.priceUsd)
}

var greenColor = color.New(color.FgWhite).Add(color.BgGreen)
var redColor = color.New(color.FgWhite).Add(color.BgRed)
var neutralColor = color.New(color.FgWhite)

func main() {
    key := os.Getenv("BFX_API_KEY")
    secret := os.Getenv("BFX_API_SECRET")
    c := bitfinex.NewClient().Auth(key, secret)

    // in case your proxy is using a non valid certificate set to TRUE
    c.WebSocketTLSSkipVerify = false

    err := c.WebSocket.Connect()
    if err != nil {
        log.Fatal("Error connecting to web socket : ", err)
    }
    defer Cleanup(c)

    usd := MyData{name: "USD", amount: 0.0, priceUsd: 1.0, color: neutralColor}
    iot := MyData{name: "IOT", amount: 0.0, priceUsd: 0.0, color: neutralColor}
    eth := MyData{name: "ETH", amount: 0.0, priceUsd: 0.0, color: neutralColor}
    xrp := MyData{name: "XRP", amount: 0.0, priceUsd: 0.0, color: neutralColor}
    prices := []MyData{usd, iot, eth, xrp}

    book_iotusd_chan := make(chan []float64)
    book_ethusd_chan := make(chan []float64)
    book_xrpusd_chan := make(chan []float64)
    c.WebSocket.AddSubscribe(bitfinex.ChanTicker, bitfinex.IOTUSD, book_iotusd_chan)
    c.WebSocket.AddSubscribe(bitfinex.ChanTicker, bitfinex.ETHUSD, book_ethusd_chan)
    c.WebSocket.AddSubscribe(bitfinex.ChanTicker, bitfinex.XRPUSD, book_xrpusd_chan)
    go ListenChanTicker(book_iotusd_chan, prices[:], 1)
    go ListenChanTicker(book_ethusd_chan, prices[:], 2)
    go ListenChanTicker(book_xrpusd_chan, prices[:], 3)

    // Private channels
    dataChan := make(chan bitfinex.TermData)
    go c.WebSocket.ConnectPrivate(dataChan)
    go ListenPrivate(dataChan, prices[:])

    // SIGINT handler
    sigint_chan := make(chan os.Signal, 1)
    signal.Notify(sigint_chan, os.Interrupt)
    go func(){
        for _ = range sigint_chan {
            Cleanup(c)
            os.Exit(0)
        }
    }()

    err = c.WebSocket.Subscribe()
    if err != nil {
        log.Fatal(err)
    }
}

func Cleanup(c *bitfinex.Client) {
    c.WebSocket.ClearSubscriptions()
    c.WebSocket.Close()
}

func ListenChanTicker(in chan []float64, prices []MyData, index int) {
    for {
        msg := <-in
        // Reset colors
        for i, _ := range prices {
            prices[i].color = neutralColor
        }
        if msg[0] > prices[index].priceUsd {
            prices[index].color = greenColor
        } else {
            prices[index].color = redColor
        }
        // Calculate totalUsd
        prices[index].priceUsd = msg[0]
        totalUsd := 0.0
        for _, p := range prices {
            totalUsd += p.priceUsd * p.amount
        }
        // Print
        for i := 1; i <= 3; i++ {
            prices[i].Print()
            fmt.Printf("    ")
        }
        fmt.Printf("~= %.1f USD\n", totalUsd)
    }
}

func ListenPrivate(in chan bitfinex.TermData, prices []MyData) {
    for {
        msg := <-in
        if strings.Contains(msg.Term, "ws") {
            for i, _ := range prices {
                if prices[i].name == msg.Data[1] {
                    prices[i].amount = msg.Data[2].(float64)
                }
            }
        }
    }
}
