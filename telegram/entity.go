package telegram

type TGMessage struct {
	Text   string    `json:"text"`
	Images []TGImage `json:"images"`
}
type TGImage struct {
	URL string `json:"url"`
}
