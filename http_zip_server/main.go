package main

// https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"http_zip_server/zipfs"
	"io"
	"net"
	"net/http"
)

func www_handle_zip(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("www_handle_zip called\n")

	zippath := "www.zip"

	z, err := zip.OpenReader(zippath)
	if err != nil {
		http.Error(res, err.Error(), 404)
		return
	}
	defer z.Close()
	io.WriteString(res, "zip!\n")
	return
	http.StripPrefix("/zip/", http.FileServer(zipfs.NewZipFS(&z.Reader))).ServeHTTP(res, req)
}

const keyServerAddr = "serverAddr"

func www_handle_root(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("www_handle_root called\n")
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()

	hasFirst := r.URL.Query().Has("first")
	first := r.URL.Query().Get("first")
	hasSecond := r.URL.Query().Has("second")
	second := r.URL.Query().Get("second")

	fmt.Printf("%s: got / request. first(%t)=%s, second(%t)=%s\n",
		ctx.Value(keyServerAddr),
		hasFirst, first,
		hasSecond, second)

	io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s: got /hello request\n", ctx.Value(keyServerAddr))
	io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", www_handle_root)
	mux.HandleFunc("/zip/", www_handle_zip)

	ctx, cancelCtx := context.WithCancel(context.Background())

	serverOne := &http.Server{
		Addr:    ":3333",
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	go func() {
		err := serverOne.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server one closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server one: %s\n", err)
		}
		cancelCtx()
	}()

	<-ctx.Done()
}
