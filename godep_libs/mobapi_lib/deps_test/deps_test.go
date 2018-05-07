package deps_test

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDepsDoesntContainExternalOnes(t *testing.T) {
	// TODO: decide if we really need this test
	t.SkipNow()

	commandGoList := exec.Command(`go`, `list`, `-f`, `{{range $dep := .Deps}}{{printf "%s\n" $dep}}{{end}}`, `../...`)
	commandGoFilterNonStandard := exec.Command(`xargs`, `go`, `list`, `-f`, `{{if not .Standard}}{{.ImportPath}}{{end}}`)
	importsListFilter := []string{
		"vendor",                   // don't need to check all dependencies of dependencies
		"motify_core_api/godep_libs/mobapi_lib",  // don't need to check unused subpackages, because it could be used externally
		"golang.org/x/net/context", // this package became non-standard in GO 1.8, standard "context" used instead
	}

	output, err := pipe(commandGoList, commandGoFilterNonStandard)
	assert.NoError(t, err)

	deps := strings.Split(string(output), "\n")
	prohibited := make([]string, 0, 5)

Loop:
	for _, dep := range deps {
		if len(dep) == 0 {
			continue Loop
		}
		for _, filter := range importsListFilter {
			if strings.Contains(dep, filter) {
				continue Loop
			}
		}
		prohibited = append(prohibited, dep)
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
