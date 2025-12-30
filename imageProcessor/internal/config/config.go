package config

type Config struct {
	StoragePath string `json:"storage_path" env:"STORAGE_PATH" required:"true"`
}
