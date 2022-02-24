package entity

import "github.com/duke-git/lancet/convertor"

type Tweets struct {
	ID        int64
	Author    string
	Tags      string
	Content   string
	Url       string
	MediaUrls string
	IsPublish int
}

func (t Tweets) MarshalBinary() ([]byte, error) {
	json, err := convertor.ToJson(t)
	return []byte(json), err
}
