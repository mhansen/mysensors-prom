package mysensors

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Load reads State from a file.
func LoadJson(f string, s interface{}) error {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, s); err != nil {
		return err
	}
	return nil
}

// Save saves state to a file.
func SaveJson(f string, s interface{}) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(f, data, os.ModePerm); err != nil {
		return err
	}
	return nil
}

// State contains some state and config that needs to be saved
// between job executions.
type State struct {
	// LastSensorID is the highest ID so far allocated to a sensor.
	LastSensorID int
}

// Load reads State from a file.
func (s *State) Load(f string) error {
	return LoadJson(f, s)
}

// Save saves state to a file.
func (s *State) Save(f string) error {
	return SaveJson(f, s)
}

type Config struct {
	// Locations maps sensor IDs to location strings.
	Locations map[string]string
}

// Load reads State from a file.
func (c *Config) Load(f string) error {
	return LoadJson(f, c)
}

// Save saves state to a file.
func (c *Config) Save(f string) error {
	return SaveJson(f, c)
}
