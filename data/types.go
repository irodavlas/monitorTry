package data

import "time"

type CatalogueItems struct {
	Items                []Item
	DominantBrand        interface{} `json:"dominant_brand"`
	SearchTrackingParams struct {
		SearchCorrelationID string `json:"search_correlation_id"`
		SearchSessionID     string `json:"search_session_id"`
	} `json:"search_tracking_params"`
	Pagination struct {
		CurrentPage  int `json:"current_page"`
		TotalPages   int `json:"total_pages"`
		TotalEntries int `json:"total_entries"`
		PerPage      int `json:"per_page"`
		Time         int `json:"time"`
	} `json:"pagination"`
	Code int `json:"code"`
}

type Item struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Region string
	Price  string `json:"price"`

	IsVisible  int         `json:"is_visible"`
	Discount   interface{} `json:"discount"`
	BrandTitle string      `json:"brand_title"`
	User       struct {
		ID         int    `json:"id"`
		Login      string `json:"login"`
		Business   bool   `json:"business"`
		ProfileURL string `json:"profile_url"`
		Photo      struct {
			ID                  int         `json:"id"`
			Width               int         `json:"width"`
			Height              int         `json:"height"`
			TempUUID            interface{} `json:"temp_uuid"`
			URL                 string      `json:"url"`
			DominantColor       string      `json:"dominant_color"`
			DominantColorOpaque string      `json:"dominant_color_opaque"`
			Thumbnails          []struct {
				Type         string      `json:"type"`
				URL          string      `json:"url"`
				Width        int         `json:"width"`
				Height       int         `json:"height"`
				OriginalSize interface{} `json:"original_size"`
			} `json:"thumbnails"`
			IsSuspicious   bool        `json:"is_suspicious"`
			Orientation    interface{} `json:"orientation"`
			HighResolution struct {
				ID          string      `json:"id"`
				Timestamp   int         `json:"timestamp"`
				Orientation interface{} `json:"orientation"`
			} `json:"high_resolution"`
			FullSizeURL string `json:"full_size_url"`
			IsHidden    bool   `json:"is_hidden"`
			Extra       struct {
			} `json:"extra"`
		} `json:"photo"`
	} `json:"user"`
	URL      string `json:"url"`
	Promoted bool   `json:"promoted"`
	Photo    struct {
		ID                  int64  `json:"id"`
		ImageNo             int    `json:"image_no"`
		Width               int    `json:"width"`
		Height              int    `json:"height"`
		DominantColor       string `json:"dominant_color"`
		DominantColorOpaque string `json:"dominant_color_opaque"`
		URL                 string `json:"url"`
		IsMain              bool   `json:"is_main"`
		Thumbnails          []struct {
			Type         string      `json:"type"`
			URL          string      `json:"url"`
			Width        int         `json:"width"`
			Height       int         `json:"height"`
			OriginalSize interface{} `json:"original_size"`
		} `json:"thumbnails"`
		HighResolution struct {
			ID          string      `json:"id"`
			Timestamp   int         `json:"timestamp"`
			Orientation interface{} `json:"orientation"`
		} `json:"high_resolution"`
		IsSuspicious bool   `json:"is_suspicious"`
		FullSizeURL  string `json:"full_size_url"`
		IsHidden     bool   `json:"is_hidden"`
		Extra        struct {
		} `json:"extra"`
	} `json:"photo"`
	FavouriteCount int         `json:"favourite_count"`
	IsFavourite    bool        `json:"is_favourite"`
	Badge          interface{} `json:"badge"`
	Conversion     interface{} `json:"conversion"`
	/*
			ServiceFee     struct {
				Amount       string `json:"amount"`
				CurrencyCode string `json:"currency_code"`
			} `json:"service_fee"`

		TotalItemPrice struct {
			Amount       string `json:"amount"`
			CurrencyCode string `json:"currency_code"`
		} `json:"total_item_price"`
	*/
	TotalItemPrice        string        `json:"total_item_price"`
	TotalItemPriceRounded interface{}   `json:"total_item_price_rounded"`
	ViewCount             int           `json:"view_count"`
	SizeTitle             string        `json:"size_title"`
	ContentSource         string        `json:"content_source"`
	Status                string        `json:"status"`
	IconBadges            []interface{} `json:"icon_badges"`
	SearchTrackingParams  struct {
		Score          float64  `json:"score"`
		MatchedQueries []string `json:"matched_queries"`
	} `json:"search_tracking_params"`
	Timestamp time.Time
}
