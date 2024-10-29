package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"rates_service/internal/models"
	"strconv"

	"github.com/opentracing/opentracing-go"
)

type (
	Depth struct {
		Timestamp int64    `json:"timestamp"`
		Asks      [1]Field `json:"asks"`
		Bids      [1]Field `json:"bids"`
	}
	Field struct {
		Price string `json:"price"`
	}
)

func parseDepth(data []byte, target *models.RatesDTO) error {
	var d Depth
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	askPrice, err := strconv.ParseFloat(d.Asks[0].Price, 32)
	if err != nil {
		return err
	}
	bidPrice, err := strconv.ParseFloat(d.Asks[0].Price, 32)
	if err != nil {
		return err
	}
	target.Timestamp = d.Timestamp
	target.AskPrice = float32(askPrice)
	target.BidPrice = float32(bidPrice)
	return nil
}

type garantex struct {
	client *http.Client
}

func NewGarantexProvider(c *http.Client) garantex {
	return garantex{client: c}
}

func (g garantex) GetRates(ctx context.Context, market string) (models.RatesDTO, error) {
	var r models.RatesDTO

	span, ctxt := opentracing.StartSpanFromContextWithTracer(ctx, opentracing.GlobalTracer(), "get depth from garantex")
	span.LogKV("result", "success")
	defer func() {
		span.Finish()
	}()

	req, err := http.NewRequestWithContext(ctxt, http.MethodGet, fmt.Sprintf("https://garantex.org/api/v2/depth?market=%s", market), nil)
	if err != nil {
		span.LogKV("result", "failed")
		return r, err
	}
	resp, err := g.client.Do(req)
	if err != nil {
		span.LogKV("result", "failed")
		return r, err
	}
	defer resp.Body.Close()
	buf := &bytes.Buffer{}
	if _, err = buf.ReadFrom(resp.Body); err != nil {
		span.LogKV("result", "failed")
		return r, err
	}
	if resp.StatusCode != http.StatusOK {
		span.LogKV("result", "failed")
		return r, fmt.Errorf("request remote https://garantex.org failed with status: %s, message: %s", resp.Status, buf.String())
	}
	if err := parseDepth(buf.Bytes(), &r); err != nil {
		span.LogKV("result", "failed")
		return r, fmt.Errorf("responce parse failed: %w", err)
	}
	return r, nil
}
