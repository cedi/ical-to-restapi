package calendar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cedi/icaltest/pkg/errors"
)

func getIcal(ctx context.Context, from string, url string) (io.ReadCloser, *errors.ResolvingError) {
	switch from {
	case "file":
		return getIcalFromFile(url)
	case "url":
		return getIcalFromURL(ctx, url)
	default:
		return nil, errors.NewResolvingError(fmt.Errorf("unsupported 'from' type"), "The only supported values for 'from' are 'file' or 'url'")
	}
}

func getIcalFromFile(path string) (io.ReadCloser, *errors.ResolvingError) {
	file, err := os.Open(path)
	return file, errors.NewResolvingError(err, "check if file path exists and is accessible")
}

func getIcalFromURL(ctx context.Context, url string) (io.ReadCloser, *errors.ResolvingError) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewResolvingError(fmt.Errorf("failed creating request for %s: %w", url, err), "")
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.NewResolvingError(fmt.Errorf("failed making request to %s: %w", url, err), "verify if URL exists and is accessible")
	}

	return resp.Body, nil
}
