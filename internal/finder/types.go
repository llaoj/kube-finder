package finder

import "time"

type ResponseFiles struct {
	Subpath string `json:"subpath"`
	Files   []File `json:"files"`
}

type File struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Mode      string    `json:"mode"`
	ModTime   time.Time `json:"modtime"`
	IsDir     bool      `json:"isdir"`
	GroupName string    `json:"groupname"`
	UserName  string    `json:"username"`
	Link      string    `json:"link"`
}
