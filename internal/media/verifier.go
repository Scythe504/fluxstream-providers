package media

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Verifier contains the client and URL configuration for checking a provider.
type Verifier struct {
	BaseURL string
	Client  *http.Client
}

// NewVerifier initializes a Verifier with a timeout.
func NewVerifier(baseURL string) *Verifier {
	return &Verifier{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// VerifyAll performs validation across all supported provider routes in sequence.
func (v *Verifier) VerifyAll(ctx context.Context) error {
	// Verify Trending and fetch a dynamic ID to test detail endpoints
	mediaList, err := v.VerifyTrending(ctx)
	if err != nil {
		return fmt.Errorf("trending validation failed: %w", err)
	}

	var testID string
	if len(mediaList) > 0 {
		testID = mediaList[0].ID
	} else {
		// Fallback to a default ID if trending is empty
		testID = "153518"
	}

	// Verify Seasonal
	if err := v.VerifySeasonal(ctx); err != nil {
		return fmt.Errorf("seasonal validation failed: %w", err)
	}

	// Verify Search
	if err := v.VerifySearch(ctx, "Frieren"); err != nil {
		return fmt.Errorf("search validation failed: %w", err)
	}

	// Verify Genre
	if err := v.VerifyGenre(ctx, "Action"); err != nil {
		return fmt.Errorf("genre validation failed: %w", err)
	}

	// Verify Airing
	if err := v.VerifyAiring(ctx); err != nil {
		return fmt.Errorf("airing validation failed: %w", err)
	}

	// Verify Schedule
	if err := v.VerifySchedule(ctx); err != nil {
		return fmt.Errorf("schedule validation failed: %w", err)
	}

	// Verify Media Details
	if err := v.VerifyMediaDetails(ctx, testID); err != nil {
		return fmt.Errorf("media details validation failed for ID %s: %w", testID, err)
	}

	// Verify Episodes
	epList, err := v.VerifyEpisodes(ctx, testID)
	if err != nil {
		return fmt.Errorf("episodes validation failed for ID %s: %w", testID, err)
	}

	var testEpNum string = "1"
	if epList != nil && len(epList.Episodes) > 0 {
		testEpNum = epList.Episodes[0].Number
	}

	// Verify Sources
	if err := v.VerifySources(ctx, testID, testEpNum); err != nil {
		return fmt.Errorf("sources validation failed for ID %s ep %s: %w", testID, testEpNum, err)
	}

	// Verify Recommendations
	if err := v.VerifyRecommendations(ctx, testID); err != nil {
		return fmt.Errorf("recommendations validation failed for ID %s: %w", testID, err)
	}

	return nil
}

// Trending
func (v *Verifier) VerifyTrending(ctx context.Context) ([]Media, error) {
	urlStr := fmt.Sprintf("%s/api/trending", v.BaseURL)
	var list []Media
	if err := v.getAndDecode(ctx, urlStr, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// Seasonal
func (v *Verifier) VerifySeasonal(ctx context.Context) error {
	urlStr := fmt.Sprintf("%s/api/seasonal?season=SUMMER&year=2023", v.BaseURL)
	var list []Media
	return v.getAndDecode(ctx, urlStr, &list)
}

// Search
func (v *Verifier) VerifySearch(ctx context.Context, query string) error {
	urlStr := fmt.Sprintf("%s/api/search?q=%s", v.BaseURL, url.QueryEscape(query))
	var list []Media
	return v.getAndDecode(ctx, urlStr, &list)
}

// Genre
func (v *Verifier) VerifyGenre(ctx context.Context, genre string) error {
	urlStr := fmt.Sprintf("%s/api/genre?genre=%s", v.BaseURL, url.QueryEscape(genre))
	var list []Media
	return v.getAndDecode(ctx, urlStr, &list)
}

// Airing
func (v *Verifier) VerifyAiring(ctx context.Context) error {
	urlStr := fmt.Sprintf("%s/api/airing", v.BaseURL)
	var list []Media
	return v.getAndDecode(ctx, urlStr, &list)
}

// Schedule
func (v *Verifier) VerifySchedule(ctx context.Context) error {
	urlStr := fmt.Sprintf("%s/api/schedule", v.BaseURL)
	var list []Episode
	return v.getAndDecode(ctx, urlStr, &list)
}

// Details
func (v *Verifier) VerifyMediaDetails(ctx context.Context, id string) error {
	urlStr := fmt.Sprintf("%s/api/%s", v.BaseURL, id)
	var m Media
	return v.getAndDecode(ctx, urlStr, &m)
}

// Episodes
func (v *Verifier) VerifyEpisodes(ctx context.Context, id string) (*EpisodeList, error) {
	urlStr := fmt.Sprintf("%s/api/%s/episodes", v.BaseURL, id)
	var el EpisodeList
	if err := v.getAndDecode(ctx, urlStr, &el); err != nil {
		return nil, err
	}
	return &el, nil
}

// Sources
func (v *Verifier) VerifySources(ctx context.Context, id, epNumber string) error {
	urlStr := fmt.Sprintf("%s/api/%s/episodes/%s/sources", v.BaseURL, id, epNumber)
	var list []Source
	return v.getAndDecode(ctx, urlStr, &list)
}

// Recommendations
func (v *Verifier) VerifyRecommendations(ctx context.Context, id string) error {
	urlStr := fmt.Sprintf("%s/api/%s/recommendations", v.BaseURL, id)
	var list []Media
	return v.getAndDecode(ctx, urlStr, &list)
}

func (v *Verifier) getAndDecode(ctx context.Context, urlStr string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.Client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 OK, got status code %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("JSON decode failed or response violates type contract: %w", err)
	}

	return nil
}

// VerifyProviderURL checks if a provider URL satisfies all type contracts.
func VerifyProviderURL(ctx context.Context, providerURL string) (string, error) {
	v := NewVerifier(providerURL)
	if err := v.VerifyAll(ctx); err != nil {
		return "", err
	}
	return "1.0.0", nil
}
