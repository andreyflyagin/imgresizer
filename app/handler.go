package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"image"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func validate(imgurl, width, height string) error {
	if imgurl == "" {
		return fmt.Errorf("url is required")
	}

	if width == "" {
		return fmt.Errorf("width is required")
	}

	if height == "" {
		return fmt.Errorf("height is required")
	}

	w, err := strconv.Atoi(width)
	if err != nil || w <= 0 {
		return fmt.Errorf("width should be a positive number")
	}

	h, err := strconv.Atoi(height)
	if err != nil || h <= 0 {
		return fmt.Errorf("height should be a positive number")
	}

	if w > maxImageWidth {
		return fmt.Errorf("max width %d limit exceeded", maxImageWidth)
	}

	if h > maxImageHeight {
		return fmt.Errorf("max height %d limit exceeded", maxImageHeight)
	}

	return nil
}

func setError(httpStatusCode int, msg string, err error, w http.ResponseWriter) {
	if err == nil {
		log.Printf("[ERROR] %s", msg)
	} else {
		log.Printf("[ERROR] %s, %v", msg, err)
	}

	w.WriteHeader(httpStatusCode)
	if _, err := fmt.Fprint(w, msg); err != nil {
		log.Printf("[ERROR] failed write message to client %v", err)
	}
}

func processImage(imgurl string, w http.ResponseWriter, width, height int) ([]byte, string) {
	client := http.Client{Timeout: time.Second * 30}
	req, err := http.NewRequest("GET", imgurl, nil)
	if err != nil {
		setError(http.StatusBadGateway, "error", err, w)
		return nil, ""
	}

	resp, err := client.Do(req)
	if err != nil {
		setError(http.StatusNotFound, "image not found", err, w)
		return nil, ""
	}

	defer func() {
		if e := resp.Body.Close(); e != nil {
			log.Printf("[ERROR] can't close body, %s", e)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		setError(resp.StatusCode, fmt.Sprintf("gateway status code: %d", resp.StatusCode), nil, w)
		return nil, ""
	}

	src, fileType, err := image.Decode(http.MaxBytesReader(w, resp.Body, maxImageSize))
	if err != nil {
		setError(http.StatusBadRequest, "failed decode or image is too big", err, w)
		return nil, ""
	}

	if fileType != "jpeg" {
		setError(http.StatusBadRequest, fmt.Sprintf("only jpeg is supported(got '%s')", fileType), err, w)
		return nil, ""
	}

	dst := imaging.Resize(src, width, height, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, dst, imaging.JPEG)
	if err != nil {
		setError(http.StatusBadRequest, "failed convert to jpeg", err, w)
	}

	var hash string

	h := sha256.New()
	_, err = h.Write(buf.Bytes())
	if err != nil {
		log.Printf("[ERROR] failed sha256 %v", err)
	} else {
		hash = hex.EncodeToString(h.Sum(nil))
	}

	return buf.Bytes(), hash
}

func getRouter() chi.Router{
	router := chi.NewRouter()
	router.Use(middleware.Throttle(500), middleware.Timeout(time.Second*60))
	router.Get("/", handler)
	return router
}

var (
	cacheService *cache
)

func handler(w http.ResponseWriter, r *http.Request) {
	imgurl := r.URL.Query().Get("url")
	widthStr := r.URL.Query().Get("width")
	heightStr := r.URL.Query().Get("height")

	if err := validate(imgurl, widthStr, heightStr); err != nil {
		setError(http.StatusBadRequest, fmt.Sprintf("failed validate: %s", err.Error()), nil, w)
		return
	}

	parsedUrl, err := url.QueryUnescape(imgurl)
	if err != nil {
		setError(http.StatusBadRequest, "url unescape error", err, w)
		return
	}

	width, err := strconv.Atoi(widthStr)
	if err != nil {
		setError(http.StatusInternalServerError, "error", err, w)
		return
	}

	height, err := strconv.Atoi(heightStr)
	if err != nil {
		setError(http.StatusInternalServerError, "error", err, w)
		return
	}

	var (
		data []byte
		hash string
	)

	cacheKey := `"` + parsedUrl + "|" + widthStr + "|" + heightStr + `"`

	if cacheService != nil {
		data, hash = cacheService.get(cacheKey)
	}

	if data == nil {
		data, hash = processImage(parsedUrl, w, width, height)
		if data != nil && cacheService != nil {
			cacheService.add(cacheKey, data, hash)
		}
	}

	if data != nil {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))

		if hash != "" {
			w.Header().Set("Etag", hash)
		}

		w.Header().Set("Cache-Control", "max-age=3600") // one hour cache
		match := r.Header.Get("If-None-Match")
		if match != "" {
			if strings.Contains(match, hash) {
				log.Printf("[INFO] not modified '%s'", hash)
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
		if _, err := w.Write(data); err != nil {
			log.Printf("[ERROR] failed write image to client %v", err)
		}
	}
}
