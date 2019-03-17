package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"image"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

func Test_Main(t *testing.T) {
	os.Args = []string{"app", "-p", "8182"}

	go func() {
		time.Sleep(300 * time.Millisecond)
		err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.Nil(t, err)
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		main()
		wg.Done()
	}()
	time.Sleep(time.Millisecond * 30)

	resp, err := http.Get("http://localhost:8182/?url=https://i.ytimg.com/vi/ktlQrO2Sifg/maxresdefault.jpg&width=100&height=100")
	require.Nil(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	_, fileType, err := image.Decode(resp.Body)
	assert.Equal(t, "jpeg", fileType)
	require.Nil(t, err)

	wg.Wait()
}
