package models

// HouseEntrance table houses_entraces
type HouseEntrance struct {
	HouseEntranceID  int      `json:"house_entrance_id"`
	EntranceType     *string  `json:"entrance_type"`
	Entrance         string   `json:"entrance"`
	Lat              *float32 `json:"lat"`
	Lon              *float32 `json:"lon"`
	Shared           *int     `json:"shared"`
	Plog             *int     `json:"plog"`
	CallerID         *string  `json:"caller_id"`
	CameraID         *int     `json:"camera_id"`
	HouseDomophoneID int      `json:"house_domophone_id"`
	DomophoneOutput  *int     `json:"domophone_output"`
	CMS              *string  `json:"cms"`
	CMSType          *int     `json:"cms_type"`
	CMSLevels        *string  `json:"cms_levels"`
	Path             *string  `json:"path"`
	Distance         *int     `json:"distance"`
	AltCameraID1     *int     `json:"alt_camera_id_1"`
	AltCameraID2     *int     `json:"alt_camera_id_2"`
	AltCameraID3     *int     `json:"alt_camera_id_3"`
	AltCameraID4     *int     `json:"alt_camera_id_4"`
	AltCameraID5     *int     `json:"alt_camera_id_5"`
	AltCameraID6     *int     `json:"alt_camera_id_6"`
	AltCameraID7     *int     `json:"alt_camera_id_7"`
}

// Domophone table houses_domophones
type Domophone struct {
	HouseDomophoneID int     `json:"house_domophone_id"`
	Enabled          int     `json:"enabled"`
	Model            string  `json:"model"`
	Server           string  `json:"server"`
	URL              string  `json:"url"`
	Credentials      string  `json:"credentials"`
	DTMF             string  `json:"dtmf"`
	FirstTime        *int    `json:"first_time"`
	NAT              *int    `json:"nat"`
	LocksAreOpen     *int    `json:"locks_are_open"`
	IP               *string `json:"ip"`
	SubID            *string `json:"sub_id"`
	Name             *string `json:"name"`
	Comments         *string `json:"comments"`
	Display          *string `json:"display"`
	Video            *string `json:"video"`
}

type Flat struct {
	HouseFlatID      int     `json:"house_flat_id" db:"house_flat_id"`
	AddressHouseID   int     `json:"address_house_id" db:"address_house_id"`
	Floor            *int    `json:"floor,omitempty" db:"floor"`
	Flat             string  `json:"flat" db:"flat"`
	Code             *string `json:"code,omitempty" db:"code"`
	Plog             *int    `json:"plog,omitempty" db:"plog"`
	ManualBlock      *int    `json:"manual_block,omitempty" db:"manual_block"`
	AutoBlock        *int    `json:"auto_block,omitempty" db:"auto_block"`
	AdminBlock       *int    `json:"admin_block,omitempty" db:"admin_block"`
	OpenCode         *string `json:"open_code,omitempty" db:"open_code"`
	AutoOpen         *int    `json:"auto_open,omitempty" db:"auto_open"`
	WhiteRabbit      *int    `json:"white_rabbit,omitempty" db:"white_rabbit"`
	SipEnabled       *int    `json:"sip_enabled,omitempty" db:"sip_enabled"`
	SipPassword      *string `json:"sip_password,omitempty" db:"sip_password"`
	LastOpened       *int    `json:"last_opened,omitempty" db:"last_opened"`
	CmsEnabled       *int    `json:"cms_enabled,omitempty" db:"cms_enabled"`
	Contract         *string `json:"contract,omitempty" db:"contract"`
	Login            *string `json:"login,omitempty" db:"login"`
	Password         *string `json:"password,omitempty" db:"password"`
	Cars             *string `json:"cars,omitempty" db:"cars"`
	SubscribersLimit *int    `json:"subscribers_limit,omitempty" db:"subscribers_limit"`
}
