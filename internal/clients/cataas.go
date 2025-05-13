package clients

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"net/url"
	"strconv"
)

type CatAAS struct {
	*BaseClient
}

func NewCatAASClient(baseURL string) *CatAAS {
	return &CatAAS{BaseClient: NewBaseClient(baseURL)}
}

func (c *CatAAS) Image(width, height int) (*image.Image, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url %q: %w", c.BaseURL, err)
	}

	q := u.Query()
	q.Set("width", strconv.Itoa(width))
	q.Set("height", strconv.Itoa(height))
	u.RawQuery = q.Encode()

	log.Printf("requesting %q", u.String())
	responseBody, err := c.DoRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request %q: %w", u.String(), err)
	}

	img, imgName, err := image.Decode(bytes.NewReader(*responseBody))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image %s: %w", imgName, err)
	}

	return &img, nil
}
