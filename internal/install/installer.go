package install

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/utils"
)

const LOGO = `
  ____              _            __     ______  _   _
 |  _ \  ___  _ __ | | _____ _   \ \   / /  _ \| \ | |
 | | | |/ _ \| '_ \| |/ / _ \ | | \ \ / /| |_) |  \| |
 | |_| | (_) | | | |   <  __/ |_| |\ V / |  __/| |\  |
 |____/ \___/|_| |_|_|\_\___|\__, | \_/  |_|   |_| \_|
                             |___/
`

var allowsEmpty = map[string]bool{
	"hosted_zone":     true,
	"vpn_domain_name": true,
}

var customPromptMessage = map[string]string{
	"subnets": " (multiple values allowed separated by a comma)",
}

var customTypes = map[string]func(string) string{
	"subnets": parseList,
}

type Wizard struct {
	BackendConfigFilename    string
	prompt                   *bufio.Reader
	ssm                      *aws.SSM
	tfvarsSettingsManager    *SettingsManager
	tfbackendSettingsManager *SettingsManager
}

func NewWizard(tfvarsTemplate, tfvarsOutput, tfbackendTemplate, tfbackendOutput string) *Wizard {
	prompt := bufio.NewReader(os.Stdin)
	ctx := context.Background()
	ssm, err := aws.NewSSM(ctx)
	if err != nil {
		fmt.Printf("Failed to create SSM client\n")
		return nil
	}

	return &Wizard{
		BackendConfigFilename: tfbackendTemplate,
		prompt:                prompt,
		ssm:                   ssm,
		tfvarsSettingsManager: &SettingsManager{
			SourceFile:      tfvarsTemplate,
			DestinationFile: tfvarsOutput,
			prompt:          prompt,
		},
		tfbackendSettingsManager: &SettingsManager{
			SourceFile:      tfbackendTemplate,
			DestinationFile: tfbackendOutput,
			prompt:          prompt,
		},
	}
}

func (w *Wizard) displayLogo() {
	fmt.Print(LOGO)
}

func (w *Wizard) readSecret(parameterName, prompt string) {
	exist := w.ssm.Exists(parameterName)
	if exist {
		if !readConfirm(w.prompt, fmt.Sprintf("%v: Value already set, do you want to replace it?", prompt)) {
			return
		}
	}

	value := readValue(w.prompt, prompt, "")
	_, err := w.ssm.SetParameter(parameterName, value, true, true)
	if err != nil {
		fmt.Printf("Could not set parameter %v: %v\n", parameterName, err.Error())
		return
	}
	fmt.Printf("Parameter %v set successfully\n", parameterName)
}

func (w *Wizard) generateSecrets() {
	printWithColors("\n\nThe next values will be saved as SSM Parameter secrets:\n")

	w.readSecret("/donkeyvpn/webhooksecret", "Webhook Secret Key")
	w.readSecret("/donkeyvpn/telegrambotapikey", "Telegram Bot API Key")
}

func (w *Wizard) generateWireguardKeys() {
	printWithColors("\nWireguard private/public key pairs\n")

	exist := w.ssm.Exists("/donkeyvpn/privatekey")
	if exist {
		if !readConfirm(w.prompt, "Wireguard priv/pub keys already generated, do you want to generate new keys?") {
			fmt.Println("Previous key pair will remain")
			return
		}
	}
	fmt.Println("Generating new Wireguard key pair...")
	keyPair, err := utils.GenerateNewKeyPair()
	if err != nil {
		fmt.Printf("Error while generating wireguard keys: %v\n", err.Error())
		return
	}
	_, err = w.ssm.SetParameter("/donkeyvpn/privatekey", *keyPair.PrivateKey, true, true)
	if err != nil {
		fmt.Printf("Failed to set /donkeyvpn/privatekey SSM parameter: %v\n", err.Error())
	}
	_, err = w.ssm.SetParameter("/donkeyvpn/publickey", *keyPair.PublicKey, true, true)
	if err != nil {
		fmt.Printf("Failed to set /donkeyvpn/publickey SSM parameter: %v\n", err.Error())
	}
	fmt.Println("Wireguard key pair created successfully")
}

func (w *Wizard) generateTfvars() {
	printWithColors("\nPlease introduce the values required by Terraform:\n\n")
	w.tfvarsSettingsManager.Process()
}

func (w *Wizard) generateTfbackend() {
	printWithColors("\nPlease introduce the values required by Terraform Backend:\n\n")
	w.tfbackendSettingsManager.Process()
}

func (w *Wizard) Start() {
	w.displayLogo()
	w.generateTfbackend()
	w.generateTfvars()
	w.generateSecrets()
	w.generateWireguardKeys()
}
