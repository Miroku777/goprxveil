package models

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Product struct {
	Id          int
	UrlId       int //OneToOne for Urls
	SellerName  string
	Name        string
	Rating      float64
	LastChecked time.Time
	Description string
	Options     json.RawMessage
}

type PriceHistory struct {
	Id        int
	ProductId int //ManyToOne
	Price     string
	CreatedAt time.Time
}

type Groups struct {
	Id     int    `json:"id"`      //OneToMany
	UserId int    `json:"user_id"` //ManyToOne
	Name   string `json:"name"`
}

// ManyToMany
type GroupUrls struct {
	GroupId int `json:"group_id"` //ManyToOne
	UrlId   int `json:"url_id"`   //ManyToOne
}

// prodact -> urls
type Urls struct {
	Id        int //OneToMany //OneToOne from Product
	Url       string
	CreatedAt time.Time
	Checked   bool
	Active    bool
}

type Users struct {
	Id       int    //OneToMany
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ChartPoint struct {
	Date  string `json:"date"`
	Price string `json:"price"`
}

type Analitics struct {
	SellerName   string          `json:"seller_name"`
	Name         string          `json:"name"`
	Rating       float64         `json:"rating"`
	Description  string          `json:"description"`
	Options      json.RawMessage `json:"options"`
	CurrentPrice string          `json:"current_price"`
	History      []ChartPoint    `json:"history"`
}

type ProductToParseUpdate struct {
	Id  int
	Url string
}

type JWTClaims struct {
	UserId int `json:"user_id"`
	jwt.RegisteredClaims
}
