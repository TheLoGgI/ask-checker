package migrate

import (
	structs "ask-checker/stucts"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type TickerPatch struct {
	Isin   string
	Ticker string
}

type NordnetSearchGroup struct {
	DisplayGroupType        string                `json:"display_group_type"`
	DisplayGroupDescription string                `json:"display_group_description"`
	Results                 []NordnetSearchResult `json:"results"`
}

type NordnetSearchResult struct {
	InstrumentID             int64             `json:"instrument_id"`
	InstrumentGroupType      string            `json:"instrument_group_type"`
	InstrumentType           string            `json:"instrument_type"`
	InstrumentTypeName       string            `json:"instrument_type_display_name"`
	InstrumentIcon           string            `json:"instrument_icon"`
	DisplayName              string            `json:"display_name"`
	DisplaySymbol            string            `json:"display_symbol"`
	DisplayNameHighlighted   string            `json:"display_name_highlighted"`
	DisplaySymbolHighlighted string            `json:"display_symbol_highlighted"`
	LastPrice                NordnetPrice      `json:"last_price"`
	ClosePrice               NordnetPrice      `json:"close_price"`
	Spread                   NordnetPrice      `json:"spread"`
	SpreadPct                float64           `json:"spread_pct"`
	Currency                 string            `json:"currency"`
	PriceUnit                string            `json:"price_unit"`
	DiffPctOneDay            float64           `json:"diff_pct_one_day"`
	DiffPctOneYear           float64           `json:"diff_pct_one_year"`
	TickTimestamp            int64             `json:"tick_timestamp"`
	ExchangeCountry          string            `json:"exchange_country"`
	Turnover                 float64           `json:"turnover"`
	TurnoverVolume           int64             `json:"turnover_volume"`
	MarketInfo               NordnetMarketInfo `json:"market_info"`
	StatusInfo               NordnetStatusInfo `json:"status_info"`
	NNXInstrumentID          string            `json:"nnx_instrument_id"`
	TradingOrderBookID       string            `json:"trading_order_book_id"`
	MarketDataOrderBookID    string            `json:"market_data_order_book_id"`
	DisplaySlug              string            `json:"display_slug"`
	InstrumentClass          string            `json:"instrument_class"`
}

type NordnetPrice struct {
	Price    float64 `json:"price"`
	Decimals int     `json:"decimals"`
}

type NordnetMarketInfo struct {
	MarketID    int64  `json:"market_id"`
	MarketSubID int64  `json:"market_sub_id"`
	Identifier  string `json:"identifier"`
	TickSizeID  int64  `json:"tick_size_id"`
}

type NordnetStatusInfo struct {
	TradingStatus           string `json:"trading_status"`
	TranslatedTradingStatus string `json:"translated_trading_status"`
	TickTimestamp           int64  `json:"tick_timestamp"`
}

func CreateDatabaseTickerPatches(ISINList []structs.Ask) []TickerPatch {
	patches := make([]TickerPatch, 0, len(ISINList))

	for _, item := range ISINList {
		if !validateISIN(item.Isin) {
			continue
		}

		groups, err := queryNordnetAPI(item.Isin)
		if err != nil {
			log.Fatal(err)
			continue
		}

		ticker := extractTicker(groups)
		if ticker == "" {
			continue
		}

		patches = append(patches, TickerPatch{
			Isin:   item.Isin,
			Ticker: ticker,
		})
	}

	return patches
}

func extractTicker(groups []NordnetSearchGroup) string {
	for _, group := range groups {
		for _, result := range group.Results {
			ticker := strings.TrimSpace(result.DisplaySymbol)
			if ticker != "" {
				return ticker
			}
		}
	}

	return ""
}

func queryNordnetAPI(isin string) ([]NordnetSearchGroup, error) {

	if isin == "" {
		return nil, fmt.Errorf("isin cannot be empty")
	}

	url := fmt.Sprintf("https://www.nordnet.dk/api/2/main_search?query=%s&search_space=ALL&limit=6", isin)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("cache-control", "no-cache")
	req.Header.Add("client-id", "NEXT")
	req.Header.Add("dnt", "1")
	req.Header.Add("ntag", "NO_NTAG_RECEIVED_YET")
	req.Header.Add("pragma", "no-cache")
	req.Header.Add("priority", "u=1, i")
	req.Header.Add("referer", "https://www.nordnet.dk/")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-fetch-site", "same-origin")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("nordnet api returned status %d", res.StatusCode)
	}

	var parsed []NordnetSearchGroup
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func validateISIN(isin string) bool {
	// ISIN format: 2 letters country code + 9 alphanumeric chars + 1 check digit.
	re := regexp.MustCompile(`^[A-Z]{2}[A-Z0-9]{9}[0-9]$`)
	return re.MatchString(isin)
}
