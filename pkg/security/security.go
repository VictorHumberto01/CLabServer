package security

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsCommandAvailable checks if a command is available in the system PATH
func IsCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// PromptForUnsecureMode prompts the user to accept running in unsecure mode
func PromptForUnsecureMode() bool {
	fmt.Println("⚠️  Warning: Firejail is not available on this system.")
	fmt.Println("This means the server will run without sandboxing, which could be a security risk.")
	fmt.Print("Do you want to continue in unsecure mode? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
