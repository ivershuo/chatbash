package main

import (
	"errors"
	"os/exec"
	"regexp"
)

func ParseBash(text string) (string, error) {
	re := regexp.MustCompile("```bash\n(.*?)\n```")
	if matches := re.FindStringSubmatch(text); len(matches) > 0 {
		return matches[1], nil
	}
	return "", errors.New("no valid bash command")
}

// 待优化
func ExecBash(bash string) (string, error) {
	cmd := exec.Command("bash", "-c", bash)
	if output, err := cmd.Output(); err == nil {
		return string(output), nil
	} else {
		return "", err
	}
}
