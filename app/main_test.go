package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"image"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

func mockGateway() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "image/jpeg")
		f, err := os.Open("../testdata/flowers1.jpg")
		if err!= nil {
			panic(err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()

		_, err = io.Copy(w, f)
		if err != nil {
			panic(err)
		}
	}))
}

func Test_main(t *testing.T) {
	os.Args = []string{"app", "-p", "8182"}

	go func() {
		time.Sleep(300 * time.Millisecond)
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.Nil(t, err)
	}()

	serv := mockGateway()
	defer serv.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		main()
		wg.Done()
	}()
	time.Sleep(time.Millisecond * 30)

	resp, err := http.Get(fmt.Sprintf("http://localhost:8182/?url=%s&width=100&height=100", serv.URL))
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	_, fileType, err := image.Decode(resp.Body)
	assert.Equal(t, "jpeg", fileType)
	require.Nil(t, err)

	wg.Wait()
}
