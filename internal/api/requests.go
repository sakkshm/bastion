package api

type JobExecRequest struct {
	Cmd []string `json:"cmd"`
}

type UploadMetadata struct {
	Path string `json:"path"`
}

type DownloadRequest struct {
	Path string `json:"path"`
}

type DeleteRequest struct {
	Path string `json:"path"`
}

type ListFileRequest struct {
	Path string `json:"path"`
	Page string `json:"page"`
	Limit string `json:"limit"`
}
