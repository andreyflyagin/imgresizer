package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_validate(t *testing.T) {
	err := validate("http://ya.ru", "200", "300")
	assert.Nil(t, err)

	err = validate("", "200", "300")
	assert.True(t, strings.Contains(err.Error(), "url is required"))

	err = validate("http://ya.ru", "", "300")
	assert.True(t, strings.Contains(err.Error(), "width is required"))

	err = validate("http://ya.ru", "100", "")
	assert.True(t, strings.Contains(err.Error(), "height is required"))

	err = validate("http://ya.ru", "-1", "300")
	assert.True(t, strings.Contains(err.Error(), "width should be a positive number"))

	err = validate("http://ya.ru", "200", "0")
	assert.True(t, strings.Contains(err.Error(), "height should be a positive number"))

	err = validate("http://ya.ru", "1001", "10")
	assert.True(t, strings.Contains(err.Error(), "max width 1000 limit exceeded"))

	err = validate("http://ya.ru", "19", "1002")
	assert.True(t, strings.Contains(err.Error(), "max height 1000 limit exceeded"))
}

func Test_handler(t *testing.T) {
	serv := mockGateway()
	defer serv.Close()

	ts := httptest.NewServer(getRouter())
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf(ts.URL + "/?url=%s/&width=100&height=100", serv.URL))
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
