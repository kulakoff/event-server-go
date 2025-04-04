package models

type Stream struct {
	ID     int
	UrlDVR string
	UrlFRS string
}

// FIXME:
type Camera struct {
	CameraID    int      `json:"camera_id"`
	Enabled     int      `json:"enabled"`
	Model       string   `json:"model"`
	URL         string   `json:"url"`
	Stream      *string  `json:"stream"`
	Credentials string   `json:"credentials"`
	Name        *string  `json:"name"`
	DVRStream   *string  `json:"dvr_stream"`
	Timezone    *string  `json:"timezone"`
	Lat         *float32 `json:"lat"`
	Lon         *float32 `json:"lon"`
	Direction   *float32 `json:"direction"`
	Angle       *float32 `json:"angle"`
	Distance    *float32 `json:"distance"`
	FRS         *string  `json:"frs"`
	Common      *int     `json:"common"`
	IP          *string  `json:"ip"`
	SubID       *string  `json:"sub_id"`
	Sound       int      `json:"sound"`
	Comments    *string  `json:"comments"`
	MdArea      *string  `json:"md_area"`
	RcArea      *string  `json:"rc_area"`
	FrsMode     *int     `json:"frs_mode"`
	Ext         *string  `json:"ext"`
}
