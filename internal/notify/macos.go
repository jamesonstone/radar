package notify

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Notifier struct{}

func New() *Notifier { return &Notifier{} }

func (n *Notifier) Notify(title, message string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	safeTitle := strings.ReplaceAll(title, `"`, `\\"`)
	safeMessage := strings.ReplaceAll(message, `"`, `\\"`)
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s"`, safeMessage, safeTitle))
	return cmd.Run()
}
