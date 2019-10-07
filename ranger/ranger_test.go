package ranger

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func wrapIO(cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stdout, stdout)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stderr, stderr)

	return nil
}

func testBuildPackage() error {
	fmt.Println("----> BEGIN GENERATION")

	rc, err := ParseFile("testdata/good.yml")
	if err != nil {
		return err
	}

	code, err := rc.GenerateCode()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile("testdata/pkg/generated.go", code, 0644); err != nil {
		return err
	}

	tests, err := rc.GenerateTest()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile("testdata/pkg/generated_test.go", tests, 0644); err != nil {
		return err
	}

	fuzz, err := rc.GenerateFuzz()
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile("testdata/pkg/generated_fuzz.go", fuzz, 0644); err != nil {
		return err
	}

	fmt.Println("----> BEGIN COMPILE")
	defer fmt.Println("----> END COMPILE")

	cmd := exec.Command("go", "build", "-mod=vendor", "./testdata/pkg")
	if err := wrapIO(cmd); err != nil {
		return err
	}
	return cmd.Run()
}

func runTestPackage() error {
	if err := testBuildPackage(); err != nil {
		return errors.Wrap(err, "building")
	}

	fmt.Println("----> BEGIN TEST")
	defer fmt.Println("----> END TEST")
	cmd := exec.Command("go", "test", "-coverprofile", "ranger.out", "-mod=vendor", "-race", "-v", "./testdata/pkg", "-count", "1")
	if err := wrapIO(cmd); err != nil {
		return err
	}

	return cmd.Run()
}

func TestBuild(t *testing.T) {
	assert.Nil(t, testBuildPackage())
}

func TestPackage(t *testing.T) {
	assert.Nil(t, runTestPackage())
}

// this table was taken from go's encoding/binary varint tests
var tests = []int64{
	-1 << 63,
	-1<<63 + 1,
	-1,
	0,
	1,
	2,
	10,
	20,
	61,
	63,
	64,
	65,
	127,
	128,
	129,
	255,
	256,
	257,
	4260323311,
	1<<63 - 1,
}

func TestUvarintSize(t *testing.T) {
	for _, x := range tests {
		buf := make([]byte, binary.MaxVarintLen64)
		i := uint64(x)
		n := binary.PutUvarint(buf, i)
		assert.Equal(t, n, UvarintSize(i), i)
		l, ni := binary.Uvarint(buf)
		assert.Equal(t, l, i, i)
		assert.Equal(t, ni, n, i)
		assert.Equal(t, ni, UvarintSize(i), i)
	}
}
