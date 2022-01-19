package model

type Resp struct {
	Code    int
	Message string
	Data    RowData
}

type RowData struct {
	Room_id      int
	Short_id     int
	Uid          int
	Live_status  int
	Playurl_info Info
}

type Info struct {
	Playurl Playurl
}

type Playurl struct {
	Stream []Stream
}

type Stream struct {
	Protocol_name string
	Format        []Format
}

type Format struct {
	Codec []Codec
}

type Codec struct {
	Current_qn int
	Base_url   string
	Url_info   []Url_info
}

type Url_info struct {
	Host  string
	Extra string
}
