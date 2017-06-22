package main

import "fmt"
import "log"
import "github.com/bitfinexcom/bitfinex-api-go/v1"
import "os"
import "os/signal"

func main() {
    c := bitfinex.NewClient()

    // in case your proxy is using a non valid certificate set to TRUE
    c.WebSocketTLSSkipVerify = false

    err := c.WebSocket.Connect()
    if err != nil {
        log.Fatal("Error connecting to web socket : ", err)
    }
    defer Cleanup(c)

    prices := []float64{0.0, 0.0, 0.0}

    book_iotusd_chan := make(chan []float64)
    book_ethusd_chan := make(chan []float64)
    book_xrpusd_chan := make(chan []float64)
    c.WebSocket.AddSubscribe(bitfinex.ChanTicker, bitfinex.IOTUSD, book_iotusd_chan)
    c.WebSocket.AddSubscribe(bitfinex.ChanTicker, bitfinex.ETHUSD, book_ethusd_chan)
    c.WebSocket.AddSubscribe(bitfinex.ChanTicker, bitfinex.XRPUSD, book_xrpusd_chan)

    go listen(book_iotusd_chan, prices[:], 0)
    go listen(book_ethusd_chan, prices[:], 1)
    go listen(book_xrpusd_chan, prices[:], 2)

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

func listen(in chan []float64, prices []float64, index int) {
    for {
        msg := <-in
        prices[index] = msg[0]
        fmt.Printf("IOTUSD: %.5f    ETHUSD: %.5f    XRPUSD: %.5f\n", prices[0], prices[1], prices[2])
    }
}
