package models

type ShortenLinkDTO struct {
	Link string `json:"link"`
}

type ShortenLinkResponseDTO struct {
	ShortURL string `json:"short_url"`
}
