package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	dolarHojeURL   = "https://dolarhoje.com/"
	euroHojeURL    = "https://dolarhoje.com/euro-hoje"
	ouroHojeURL    = "https://dolarhoje.com/ouro-hoje"
	bitcoinHojeURL = "https://dolarhoje.com/bitcoin-hoje"
	requestTimeout = 1000 * time.Millisecond
	valueRegex     = `id="nacional" value="([0-9,]+)"`
)

func fetchURLContent(url string, timeout time.Duration) ([]byte, error) {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func parseValue(content []byte) (float64, error) {
	re := regexp.MustCompile(valueRegex)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		return 0, fmt.Errorf("cotação não encontrada")
	}

	value := strings.Replace(matches[1], ",", ".", 1)
	return strconv.ParseFloat(value, 64)
}

func getQuote(url, label string, ch chan<- string, timeout time.Duration) {
	content, err := fetchURLContent(url, timeout)
	if err != nil {
		return
	}

	value, err := parseValue(content)
	if err != nil {
		return
	}

	valueStr := strings.Replace(fmt.Sprintf("%.2f", value), ".", ",", 1)
	ch <- fmt.Sprintf("%s R$ %s", label, valueStr)
}

func main() {
	urls := []string{dolarHojeURL, euroHojeURL, ouroHojeURL, bitcoinHojeURL}
	labels := []string{"Dólar Hoje", "Euro", "Ouro", "Bitcoin"}

	ch := make(chan string, len(urls))

	for i, url := range urls {
		go getQuote(url, labels[i], ch, requestTimeout)
	}

	for range urls {
		select {
		case msg := <-ch:
			fmt.Println(msg)
		case <-time.After(requestTimeout):
		}
	}
}
