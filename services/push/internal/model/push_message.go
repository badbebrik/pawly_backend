package model

type PushMessage struct {
	Title string
	Body  string
	Data  map[string]string
}
