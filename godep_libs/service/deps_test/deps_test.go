package deps_test

import (
	"github.com/stretchr/testify/assert"

	"bytes"
	"io"
	"os/exec"
	"strings"
	"testing"
)

func TestDepsDoesntContainExternalOnes(t *testing.T) {
	commandGoList := exec.Command(`go`, `list`, `-f`, `{{range $dep := .Deps}}{{printf "%s\n" $dep}}{{end}}`, `../...`)
	commandGoFilterNonStandard := exec.Command(`xargs`, `go`, `list`, `-f`, `{{if not .Standard}}{{.ImportPath}}{{end}}`)

	output, err := pipe(commandGoList, commandGoFilterNonStandard)
	assert.NoError(t, err)

	deps := strings.Split(string(output), "\n")
	prohibited := make([]string, 0, 5)
	for _, dep := range deps {
		if len(dep) > 0 && !strings.Contains(dep, `vendor`) && !strings.Contains(dep, `godep.lzd.co/service`) {
			prohibited = append(prohibited, dep)
		}

	}

	if !assert.ObjectsAreEqual([]string{}, prohibited) {
		t.Errorf("There are prohibited dependencies:\n%s", strings.Join(prohibited, "\n"))
	}
}

func pipe(c1, c2 *exec.Cmd) (output string, err error) {
	r, w := io.Pipe()

	c1.Stdout = w
	c2.Stdin = r

	var out bytes.Buffer
	c2.Stdout = &out

	c1.Start()
	c2.Start()

	c1.Wait()
	w.Close()

	c2.Wait()
	r.Close()

	return out.String(), nil
}
