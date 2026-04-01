package api

const (
	HealthEndpoint = "/health"
)

// Session Endpoint
const (
	CreateSessionEndpoint    = "/session/create"
	SessionBaseEndpoint      = "/session/{id}"
	GetSessionStatusEndpoint = "/status"
	StartSessionEndpoint     = "/start"
	StopSessionEndpoint      = "/stop"
	DeleteSessionEndpoint    = "/"
)

// Job Endpoint
const (
	JobExecuteEndpoint   = "/exec"
	GetJobStatusEndpoint = "/job/{job_id}"
)
