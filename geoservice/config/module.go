package config

func Load() (*Config, error) {
	dir, err := ReadFileEnv(".env")
	if err != nil {
		return nil, err
	}
	return loadFromConsul(dir.HostConsul, dir.KeyConsul, dir.ServiceConsul)
}
