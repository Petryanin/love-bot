package clients

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"log"
	"net/url"
	"strconv"
)

type CatGetter interface {
	Image(ctx context.Context, width, height int) (*image.Image, error)
}

type CatAASClient struct {
	api Requester
}

var _ CatGetter = (*CatAASClient)(nil)

func NewCatAASClient(baseURL string) *CatAASClient {
	return &CatAASClient{api: NewBaseClient(baseURL)}
}

func (c *CatAASClient) Image(ctx context.Context, width, height int) (*image.Image, error) {
	u, err := url.Parse(c.api.BaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %q: %w", c.api.BaseURL(), err)
	}

	q := u.Query()
	q.Set("width", strconv.Itoa(width))
	q.Set("height", strconv.Itoa(height))
	u.RawQuery = q.Encode()

	log.Printf("requesting %q", u.String())
	responseBody, err := c.api.DoRequest(ctx, "GET", u.String(), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request %q: %w", u.String(), err)
	}

	img, imgName, err := image.Decode(bytes.NewReader(responseBody))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image %s: %w", imgName, err)
	}

	return &img, nil
}
