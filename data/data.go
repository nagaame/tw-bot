package data

import "github.com/duke-git/lancet/convertor"

type Tweet struct {
	ID        int64
	Author    string
	Tags      string
	Content   string
	Url       string
	MediaUrls string
	IsPublish int
}

func (t Tweet) MarshalBinary() ([]byte, error) {
	json, err := convertor.ToJson(t)
	return []byte(json), err
}
