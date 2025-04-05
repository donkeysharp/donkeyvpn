package install

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/donkeysharp/donkeyvpn/internal/utils"
)

func printWithColors(message string) {
	NO_FORMAT := "\033[0m"
	C_GREY := "\033[38;5;237m"
	C_SPRINGGREEN2 := "\033[48;5;47m"
	fmt.Fprintf(os.Stdout, "%v%v%v%v", C_GREY, C_SPRINGGREEN2, message, NO_FORMAT)
}

func cleanValue(value string) string {
	value = strings.ReplaceAll(value, "\"", "")
	value = strings.ReplaceAll(value, "[", "")
	value = strings.ReplaceAll(value, "]", "")
	return value
}

func loadKeyValue(filename string) (*utils.OrderedMap, error) {
	inputFile, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening template vars file: %v\n", err.Error())
		return nil, err
	}
	defer inputFile.Close()

	scanner := bufio.NewScanner(inputFile)

	result := utils.NewOrderedMap()
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := ""
		if len(parts) > 1 {
			value = strings.TrimSpace(parts[1])
		}
		value = cleanValue(value)
		result.Set(key, value)
	}
	return result, nil
}

func readValue(prompt *bufio.Reader, key, defaultValue string) string {
	_, allowsEmpty := allowsEmpty[key]
	customPrompt := customPromptMessage[key]
	fmt.Printf("Read value for %v", key)
	if defaultValue != "" {
		fmt.Printf(" (%v)", defaultValue)
	}
	fmt.Printf("%v: ", customPrompt)

	value, _ := prompt.ReadString('\n')
	value = strings.TrimSpace(value)
	if value == "" && defaultValue != "" {
		return defaultValue
	} else if value == "" && allowsEmpty {
		return ""
	} else if value == "" && defaultValue == "" {
		return readValue(prompt, key, defaultValue)
	}
	return value
}

func readConfirm(prompt *bufio.Reader, message string) bool {
	fmt.Printf("%v: y/N (N default) ", message)
	value, _ := prompt.ReadString('\n')
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	return value == "y"
}

func parseList(value string) string {
	items := strings.Split(value, ",")
	result := "["
	for idx, item := range items {
		if idx != 0 {
			result += ","
		}
		item = strings.TrimSpace(item)
		result += fmt.Sprintf("\"%v\"", item)
	}
	result += "]"
	return result
}

func parseValue(key, value string) string {
	parser, exists := customTypes[key]
	if !exists {
		return fmt.Sprintf("\"%v\"", value)
	}
	return parser(value)
}
