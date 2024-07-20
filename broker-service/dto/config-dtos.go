package dto

type ServerDto struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Admin    string `json:"admin"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
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
