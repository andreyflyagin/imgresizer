package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_handler(t *testing.T) {
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
