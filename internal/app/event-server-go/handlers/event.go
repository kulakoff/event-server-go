package handlers

type Events struct {
	NotAnswered  int
	Answered     int
	OpenByKey    int
	OpenByApp    int
	OpenByFaceID int
	OpenByCode   int
	OpenByCall   int
	OpenByButton int
}

var Event Events

func init() {
	Event = Events{
		NotAnswered:  1,
		Answered:     2,
		OpenByKey:    3,
		OpenByApp:    4,
		OpenByFaceID: 5,
		OpenByCode:   6,
		OpenByCall:   7,
		OpenByButton: 8,
	}
}
