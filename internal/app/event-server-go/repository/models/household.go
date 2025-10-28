package models

// HouseEntrance table houses_entraces
type HouseEntrance struct {
	AddressHouseID   int      `json:"address_house_id"`
	Prefix           int      `json:"prefix"`
	HouseEntranceID  int      `json:"house_entrance_id"`
	EntranceType     *string  `json:"entrance_type"`
	Entrance         string   `json:"entrance"`
	Lat              *float32 `json:"lat"`
	Lon              *float32 `json:"lon"`
	Shared           *int     `json:"shared"`
	Plog             *int     `json:"plog"`
	CallerID         *string  `json:"caller_id"`
	CameraID         *int     `json:"camera_id,omitempty"`
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
	HouseFlatID      int     `json:"house_flat_id"`
	AddressHouseID   int     `json:"address_house_id"`
	Floor            *int    `json:"floor,omitempty"`
	Flat             string  `json:"flat"`
	Code             *string `json:"code,omitempty"`
	Plog             *int    `json:"plog,omitempty"`
	ManualBlock      *int    `json:"manual_block,omitempty"`
	AutoBlock        *int    `json:"auto_block,omitempty"`
	AdminBlock       *int    `json:"admin_block,omitempty"`
	OpenCode         *string `json:"open_code,omitempty"`
	AutoOpen         *int    `json:"auto_open,omitempty"`
	WhiteRabbit      *int    `json:"white_rabbit,omitempty"`
	SipEnabled       *int    `json:"sip_enabled,omitempty"`
	SipPassword      *string `json:"sip_password,omitempty"`
	LastOpened       *int    `json:"last_opened,omitempty"`
	CmsEnabled       *int    `json:"cms_enabled,omitempty"`
	Contract         *string `json:"contract,omitempty"`
	Login            *string `json:"login,omitempty"`
	Password         *string `json:"password,omitempty"`
	Cars             *string `json:"cars,omitempty"`
	SubscribersLimit *int    `json:"subscribers_limit,omitempty"`
}

type RFID struct {
	HouseRfidId int     `json:"house_rfid_id"`
	RFID        string  `json:"rfid"`
	AccessType  int     `json:"access_type"`
	AccessTo    int     `json:"access_to"`
	LastSeen    *int    `json:"last_seen"`
	Comments    *string `json:"comments"`
	//Watch       int     `json:"watch"`
}

type Watcher struct {
	WatcherID   int    `json:"house_watcher_id"`
	DeviceID    int    `json:"subscriber_device_id"`
	FlatID      int    `json:"house_flat_id"`
	EventType   string `json:"event_type"` // int or str
	EventDetail string `json:"event_detail"`
	Comments    string `json:"comments"`
}

type MobileDevice struct {
	DeviceID      int    `json:"subscriber_device_id"`
	SubscriberID  int    `json:"house_subscriber_id"`
	DeviceToken   string `json:"device_token"`
	AuthToken     string `json:"auth_token"`
	Platform      int    `json:"platform"`
	PushToken     string `json:"push_token"`
	PushTokenType int    `json:"push_token_type"`
	VoipToken     string `json:"voip_token"`
	Registered    int    `json:"registered"`
	LastSeen      int    `json:"last_seen"`
	VoipEnabled   int    `json:"voip_enabled"`
	UA            string `json:"ua"`
	IP            string `json:"ip"`
	PushDisable   int    `json:"push_disabled"`
	MoneyDisable  int    `json:"money_disabled"`
	Version       string `json:"version"`
	Bundle        string `json:"bundle"`
}

type House struct {
	HouseID       int      `json:"address_house_id"`
	SettlementID  *int     `json:"address_settlement_id,omitempty"`
	StreetID      int      `json:"address_street_id"`
	HouseUUID     string   `json:"house_uuid"`
	HouseType     string   `json:"house_type"`
	HouseTypeFull string   `json:"house_type_full"`
	HouseFull     string   `json:"house_full"`
	House         string   `json:"house"`
	Lat           *float64 `json:"lat,omitempty"`
	Lon           *float64 `json:"lon,omitempty"`
	CompanyID     int      `json:"company_id"`
}
