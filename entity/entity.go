package entity

type Tweets struct {
	ID        int64
	Author    string
	Tags      string
	Content   string
	Url       string
	MediaUrls string
	IsPublish int
}
