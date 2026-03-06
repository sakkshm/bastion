package api

const (
	HealthEndpoint = "/health"

	CreateSessionEndpoint    = "/session/create"
	SessionBaseEndpoint      = "/session/{id}"
	StartSessionEndpoint     = "/start"
	StopSessionEndpoint      = "/stop"
	DeleteSessionEndpoint    = "/"
	GetSessionStatusEndpoint = "/status"
	SessionExecuteEndpoint   = "/exec"
	GetSessionLogsEndpoint   = "/logs"
)
