package dto

type ServerDto struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Admin    string `json:"admin"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

type ServerResponseDto struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type SetGrafanaDto struct {
	GrafanaUID string `json:"grafana_uid"`
}

type DatabaseDto struct {
	Name            string `json:"name"`
	Password        string `json:"password"`
	Server          string `json:"server"`
	Environment     string `json:"environment"`
	ServiceType     string `json:"service_type"`
	ComputeType     string `json:"compute_type"`
	MaxStorageSize  string `json:"max_storage_size"`
	StorageSizeUnit string `json:"storage_size_unit"`
	Connectivity    string `json:"connectivity"`
	Type            string `json:"type"`
	Version         string `json:"version"`
	Email           string `json:"email,omitempty"`
}

type ConfigurationDto struct {
	ServiceType     string `json:"service_type"`
	ComputeType     string `json:"compute_type"`
	MaxStorageSize  string `json:"max_storage_size"`
	StorageSizeUnit string `json:"storage_size_unit"`
}

type NodeDatabaseDto struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	User     string `json:"user"`
}

type DatabaseOverviewDto struct {
	Status        string           `json:"status"`
	Location      string           `json:"location"`
	Server        string           `json:"server"`
	Environment   string           `json:"environment"`
	Connectivity  string           `json:"connectivity"`
	Configuration ConfigurationDto `json:"configuration"`
	Type          string           `json:"type"`
	Version       string           `json:"version"`
	NodeIP        string           `json:"node_ip"`
	NodePort      string           `json:"node_port"`
}

type DatabaseGrafanaDto struct {
	GrafanaUID    string `json:"grafana_uid"`
	DirectoryUUID string `json:"directory_uuid"`
}

type ResourceDto struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Server   string `json:"server"`
}
