package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan string)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			// return one of these status code at random
			resp, err := makeRequest(ctx,
				"http://httpbin.org/status/200,200,200,500")
			if err != nil {
				fmt.Println("error in status goroutine:", err)
				cancel()
				return
			}
			if resp.StatusCode == http.StatusInternalServerError {
				fmt.Println("bad status, exiting")
				cancel()
				return
			}
			select {
			case ch <- "success from status":
			case <-ctx.Done():
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		defer wg.Done()
		for {
			// return after a 1 second delay
			resp, err := makeRequest(ctx, "http://httpbin.org/delay/1")
			if err != nil {
				fmt.Println("error in delay goroutine:", err)
				cancel()
				return
			}
			select {
			case ch <- "success from delay: " + resp.Header.Get("date"):
			case <-ctx.Done():
				fmt.Println("Need to cancel")
				return
			}
		}
	}()

loop:
	for {
		select {
		case s := <-ch:
			fmt.Println("in main:", s)
		case <-ctx.Done():
			fmt.Println("in main: cancelled!")
			break loop
		}
	}
	wg.Wait()
}
