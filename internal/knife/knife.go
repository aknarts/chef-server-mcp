package knife

import "os/exec"

// RunKnifeCommand executes a knife command and returns its output
func RunKnifeCommand(args ...string) (string, error) {
	cmd := exec.Command("knife", args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}
