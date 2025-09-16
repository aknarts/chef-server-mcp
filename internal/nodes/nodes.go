package nodes

import (
	"errors"
	"strings"
)

// ListNodes consolidates logic for retrieving nodes via Chef API first, then knife fallback if enabled.
// chefClient may be nil. runKnife must execute the knife command and return its combined output.
func ListNodes(chefClient interface{ ListNodes() ([]string, error) }, knifeFallback bool, runKnife func(args ...string) (string, error)) ([]string, error) {
	result := []string{}
	var apiErr error
	if chefClient != nil {
		if list, err := chefClient.ListNodes(); err == nil && list != nil {
			result = list
		} else if err != nil {
			apiErr = err
		}
	}
	if len(result) == 0 && knifeFallback {
		out, err := runKnife("node", "list")
		if err != nil {
			msg := "knife failed"
			if apiErr != nil {
				msg += "; apiErr=" + apiErr.Error()
			}
			msg += " knifeErr=" + err.Error()
			return nil, errors.New(msg)
		}
		for _, line := range splitLines(out) {
			if line != "" {
				result = append(result, line)
			}
		}
	}
	return result, nil
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(strings.TrimSpace(s), "\n")
}
