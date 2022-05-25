package main

import (
	"context"
	"net/http"

	"github.com/lidofinance/go-template/internal/app/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := http.NewServeMux()

	if err := server.RunHTTPServer(ctx, 8080, mux); err != nil {
		println(err)
	}
}
