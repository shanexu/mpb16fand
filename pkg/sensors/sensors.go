package sensors

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TempSensor temperature sensor
type TempSensor struct {
	LabelPath string
	InputPath string
	Value     int
	Name      string
	CatAt     time.Time
}

// FanSensor fan sensor
type FanSensor struct {
	LabelPath  string
	InputPath  string
	ManualPath string
	MaxPath    string
	MinPath    string
	OutputPath string
	SafePath   string
	Name       string
	Input      int
	Manual     int
	Max        int
	Min        int
	Output     int
	Safe       int
	CatAt      time.Time
}

func (s *TempSensor) UpdateValue(t time.Time) error {
	value, err := readIntFromFile(s.InputPath)
	if err != nil {
		return err
	}
	s.CatAt = t
	s.Value = value
	return nil
}

func NewTempSensor(labelPath string) (*TempSensor, error) {
	inputPath := labelPath[:len(labelPath)-5] + "input"
	b1, err := ioutil.ReadFile(labelPath)
	if err != nil {
		return nil, err
	}
	name := strings.Trim(string(b1), "\n")
	b2, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}
	value, err := strconv.Atoi(strings.Trim(string(b2), "\n"))
	if err != nil {
		return nil, err
	}
	catAt := time.Now()
	return &TempSensor{
		LabelPath: labelPath,
		InputPath: inputPath,
		Value:     value,
		Name:      name,
		CatAt:     catAt,
	}, nil
}

func readIntFromFile(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.Trim(string(b), "\n"))
}

func NewFanSensor(labelPath string) (*FanSensor, error) {
	catAt := time.Now()
	base := filepath.Base(labelPath)
	name := base[:len(base)-6]
	pathPrefix := labelPath[:len(labelPath)-5]
	inputPath := pathPrefix + "input"
	input, err := readIntFromFile(inputPath)
	if err != nil {
		return nil, err
	}
	manualPath := pathPrefix + "manual"
	manual, err := readIntFromFile(manualPath)
	if err != nil {
		return nil, err
	}
	maxPath := pathPrefix + "max"
	max, err := readIntFromFile(maxPath)
	if err != nil {
		return nil, err
	}
	minPath := pathPrefix + "min"
	min, err := readIntFromFile(minPath)
	if err != nil {
		return nil, err
	}
	outputPath := pathPrefix + "output"
	output, err := readIntFromFile(outputPath)
	if err != nil {
		return nil, err
	}
	safePath := pathPrefix + "safe"
	safe, err := readIntFromFile(safePath)
	if err != nil {
		return nil, err
	}
	return &FanSensor{
		LabelPath:  labelPath,
		InputPath:  inputPath,
		Input:      input,
		ManualPath: manualPath,
		Manual:     manual,
		MaxPath:    maxPath,
		Max:        max,
		MinPath:    minPath,
		Min:        min,
		OutputPath: outputPath,
		Output:     output,
		SafePath:   safePath,
		Safe:       safe,
		Name:       name,
		CatAt:      catAt,
	}, nil
}

func (s *FanSensor) UpdateValue(t time.Time) error {
	input, err := readIntFromFile(s.InputPath)
	if err != nil {
		return err
	}
	s.Input = input
	manual, err := readIntFromFile(s.ManualPath)
	if err != nil {
		return err
	}
	s.Manual = manual
	max, err := readIntFromFile(s.MaxPath)
	if err != nil {
		return err
	}
	s.Max = max
	min, err := readIntFromFile(s.MinPath)
	if err != nil {
		return err
	}
	s.Min = min
	output, err := readIntFromFile(s.OutputPath)
	if err != nil {
		return err
	}
	s.Output = output
	safe, err := readIntFromFile(s.SafePath)
	if err != nil {
		return err
	}
	s.Safe = safe
	s.CatAt = t
	return nil
}

var ErrDuplicatedName = errors.New("duplicated name")

func addTempSensor(s *TempSensor) error {
	if _, ok := TempSensors[s.Name]; ok {
		return ErrDuplicatedName
	}
	TempSensors[s.Name] = s
	return nil
}

func addFanSensor(s *FanSensor) error {
	if _, ok := FanSensors[s.Name]; ok {
		return ErrDuplicatedName
	}
	FanSensors[s.Name] = s
	return nil
}

var TempSensors map[string]*TempSensor = make(map[string]*TempSensor)

var FanSensors map[string]*FanSensor = make(map[string]*FanSensor)

func init() {
	filepath.Walk("/sys/devices", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasPrefix(base, "temp") && strings.HasSuffix(base, "_label") {
			s, err := NewTempSensor(path)
			if err != nil {
				panic(err)
			}
			if err := addTempSensor(s); err != nil {
				panic(err)
			}
		} else if strings.HasPrefix(base, "fan") && strings.HasSuffix(base, "_label") {
			s, err := NewFanSensor(path)
			if err != nil {
				panic(err)
			}
			if err := addFanSensor(s); err != nil {
				panic(err)
			}
		}
		return nil
	})
	log.Println("collect", len(TempSensors), "temp sensors")
	log.Println("collect", len(FanSensors), "fan sensors")

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for {
			t := <-ticker.C
			for name, sensor := range TempSensors {
				if err := sensor.UpdateValue(t); err != nil {
					log.Println("update value for", name, "failed", "with error", err)
				}
			}
			for name, sensor := range FanSensors {
				if err := sensor.UpdateValue(t); err != nil {
					log.Println("update value for", name, "failed", "with error", err)
				}
			}
		}
	}()
}
