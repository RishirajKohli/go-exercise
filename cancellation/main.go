package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
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
				cancel(fmt.Errorf("in random goroutine %w", err))
				return
			}
			if resp.StatusCode == http.StatusInternalServerError {
				fmt.Println("bad status, exiting")
				cancel(errors.New("Bad status"))
				return
			}
			select {
			case ch <- "success from status":
			case <-ctx.Done():
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
				cancel(fmt.Errorf("In delay goroutine %w", err))
				return
			}
			select {
			case ch <- "success from delay: " + resp.Header.Get("date"):
			case <-ctx.Done():
			}
		}
	}()

loop:
	for {
		select {
		case s := <-ch:
			fmt.Println("in main:", s)
		case <-ctx.Done():
			fmt.Println("in main: cancelled!", context.Cause(ctx))
			break loop
		}
	}
	wg.Wait()
}
