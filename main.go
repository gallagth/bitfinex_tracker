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

    book_iotusd_chan := make(chan []float64)

    c.WebSocket.AddSubscribe(bitfinex.ChanTicker, bitfinex.IOTUSD, book_iotusd_chan)

    go listen(book_iotusd_chan, "IOTUSD:")

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

func listen(in chan []float64, message string) {
    for {
        msg := <-in
        fmt.Println(message, msg[0])
    }
}
