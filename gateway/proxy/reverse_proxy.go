package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type ReverseProxy struct {
	userServiceURL  string
	orderServiceURL string
	client          *http.Client
}

func NewReverseProxy(userServiceURL, orderServiceURL string) *ReverseProxy {
	return &ReverseProxy{
		userServiceURL:  userServiceURL,
		orderServiceURL: orderServiceURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *ReverseProxy) ProxyToUserService(w http.ResponseWriter, r *http.Request) {
	p.proxy(w, r, p.userServiceURL)
}

func (p *ReverseProxy) ProxyToOrderService(w http.ResponseWriter, r *http.Request) {
	p.proxy(w, r, p.orderServiceURL)
}

func (p *ReverseProxy) proxy(w http.ResponseWriter, r *http.Request, targetURL string) {
	// Создаём URL для целевого сервиса
	target, err := url.Parse(targetURL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "PROXY_ERROR", "failed to parse target URL")
		return
	}

	// Копируем оригинальный путь и query параметры
	target.Path = r.URL.Path
	target.RawQuery = r.URL.RawQuery

	// Читаем тело запроса
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, err = io.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to read request body")
			return
		}
		r.Body.Close()
	}

	// Создаём новый запрос
	proxyReq, err := http.NewRequest(r.Method, target.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "PROXY_ERROR", "failed to create proxy request")
		return
	}

	// Копируем все заголовки
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Добавляем X-Forwarded-For
	proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
	proxyReq.Header.Set("X-Forwarded-Host", r.Host)
	proxyReq.Header.Set("X-Forwarded-Proto", r.URL.Scheme)

	// Логируем запрос
	log.Printf("Proxying %s %s → %s", r.Method, r.URL.Path, target.String())

	// Выполняем запрос
	resp, err := p.client.Do(proxyReq)
	if err != nil {
		log.Printf("Proxy error: %v", err)
		respondWithError(w, http.StatusBadGateway, "SERVICE_UNAVAILABLE", "target service is unavailable")
		return
	}
	defer resp.Body.Close()

	// Копируем заголовки ответа
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Устанавливаем статус код
	w.WriteHeader(resp.StatusCode)

	// Копируем тело ответа
	io.Copy(w, resp.Body)
}

func respondWithError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
