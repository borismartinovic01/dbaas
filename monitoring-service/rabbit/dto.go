package rabbit

type Job struct {
	DashboardName string
	DbType        string
	NodeIP        string
	NodePort      string
	Datname       string
}

type SetGrafanaDto struct {
	GrafanaUID string `json:"grafana_uid"`
}
