package filesystem

import "time"

type FileEntry struct {
    Name    string    `json:"name"`
    IsDir   bool      `json:"is_dir"`
    Size    int64     `json:"size,omitempty"`
    Mode    string    `json:"mode,omitempty"`
    ModTime time.Time `json:"mod_time,omitempty"`
}