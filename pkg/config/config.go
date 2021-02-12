package config

import ()

type Config struct {
	ActiveProfile string             `toml:"active_profile"`
	Profiles      map[string]Profile `toml:"profiles"`
}

type Profile struct {
	Manual   bool                `toml:"manual"`
	Speed    int                 `toml:"speed"`
	HighTemp int                 `toml:"high_temp"`
	LowTemp  int                 `toml:"low_temp"`
	Profiles map[string]*Profile `toml:"profiles"`
}
