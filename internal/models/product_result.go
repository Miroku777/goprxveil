package models

import "encoding/json"

// parse card.json
type CardResponse struct {
	NmId        int             `json:"nm_id"`
	Description string          `json:"description"`
	Options     json.RawMessage `json:"options"`
}

// parse details.json for WB
type DetailResponse struct {
	Products []struct {
		SellerName string  `json:"supplier"`
		Name       string  `json:"name"`
		Rating     float64 `json:"reviewRating"`
		Sizes      []struct {
			Price struct {
				Product int64 `json:"product"`
			} `json:"price"`
		} `json:"sizes"`
	} `json:"products"`
}

// parse details.json for ozon
type OzonJSONLD struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Rating      struct {
		Value string `json:"ratingValue"`
	} `json:"aggregateRating"`
	Offers struct {
		Price string `json:"price"`
	} `json:"offers"`
}
