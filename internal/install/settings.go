package install

import (
	"bufio"
	"fmt"
	"os"

	"github.com/donkeysharp/donkeyvpn/internal/utils"
)

type SettingsManager struct {
	SourceFile       string
	DestinationFile  string
	existingSettings *utils.OrderedMap
	baseSettings     *utils.OrderedMap
	prompt           *bufio.Reader
}

func (s *SettingsManager) loadExistingSettings() error {
	settings := utils.NewOrderedMap()
	var err error
	if utils.FileExists(s.DestinationFile) {
		settings, err = loadKeyValue(s.DestinationFile)
		if err != nil {
			return nil
		}

	}
	s.existingSettings = settings

	return nil
}

func (s *SettingsManager) loadBaseSettings() error {
	settings, err := loadKeyValue(s.SourceFile)
	if err != nil {
		return err
	}
	s.baseSettings = settings

	return nil
}

func (s *SettingsManager) saveSettings(settings *utils.OrderedMap) error {
	message := ""
	for _, key := range settings.Keys() {
		rawValue, _ := settings.Get(key)
		message += fmt.Sprintf("%v=%v\n", key, rawValue)
	}
	outputFile, err := os.Create(s.DestinationFile)
	defer outputFile.Close()

	if err != nil {
		fmt.Printf("Failed to create file %v: %v", s.DestinationFile, err.Error())
		return err
	}
	writer := bufio.NewWriter(outputFile)
	_, err = writer.WriteString(message)

	if err != nil {
		fmt.Printf("Error writing to %v\n", s.DestinationFile)
		return err
	}
	err = writer.Flush()
	if err != nil {
		fmt.Printf("Error flushing changes to disk\n")
		return err
	}

	return nil
}

func (s *SettingsManager) Process() {
	if err := s.loadExistingSettings(); err != nil {
		fmt.Printf("Failed to load existing settings: %v", err.Error())
		return
	}

	if err := s.loadBaseSettings(); err != nil {
		fmt.Printf("Failed to load base settings: %v", err.Error())
		return
	}

	finalSettings := utils.NewOrderedMap()
	for _, key := range s.baseSettings.Keys() {
		defaultValue, _ := s.baseSettings.Get(key)

		if _, exists := s.existingSettings.Get(key); exists {
			defaultValue, _ = s.existingSettings.Get(key)
		}

		value := readValue(s.prompt, key, defaultValue)
		value = parseValue(key, value)

		finalSettings.Set(key, value)
	}
	err := s.saveSettings(finalSettings)
	if err != nil {
		fmt.Printf("Failed to save settings: %v\n", err.Error())
		return
	}
	fmt.Printf("Settings written successfully to %v\n", s.DestinationFile)
}
