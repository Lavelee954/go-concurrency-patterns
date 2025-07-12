// This implemented a sample from the video
// https://www.youtube.com/watch?v=LSzR0VEraWw
package main

import (
	"context"
	"log"
	"os"
	"time"
)

func sleepAndTalk(ctx context.Context, d time.Duration, msg string) {
	select {
	case <-time.After(d):
		log.Println(msg)
	case <-ctx.Done():
		log.Println(ctx.Err())
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		server()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "client" {
		client()
		return
	}

	// Default: run the sleepAndTalk example
	log.Println("started")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	time.AfterFunc(time.Second, cancel)
	sleepAndTalk(ctx, 5*time.Second, "hello")
}
