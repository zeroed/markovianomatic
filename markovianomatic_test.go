package markovianomatic_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/zeroed/markovianomatic"
)

func buildChain(n int) *markovianomatic.Chain {
	var aChain = markovianomatic.NewChain(n)
	b := bytes.NewBufferString("The quick fox jumps over the lazy dog")
	aChain.Build(b)
	return aChain
}

func buildAChain() *markovianomatic.Chain {
	return buildChain(2)
}

func TestNewChain(t *testing.T) {
	var aChain = markovianomatic.NewChain(2)
	if l := aChain.Length(); l != 0 {
		t.Errorf("Chain Length is expected to be 0 but was %d", l)
	}
}

func TestString(t *testing.T) {
	desired := `[ ]: the
[ the]: quick
[fox jumps]: over
[jumps over]: the
[over the]: lazy
[quick fox]: jumps
[the lazy]: dog
[the quick]: fox`

	aChain := buildAChain()
	printed := aChain.String()
	if desired != printed {
		t.Errorf("Chain was expected to print %s but instead printed %s", desired, printed)
	}
}

func TestLength(t *testing.T) {
	aChain := buildAChain()
	if l := aChain.Length(); l != 8 {
		t.Errorf("Chain was exptected to be long 8 Prefixes but insted is long %d", l)
	}
}

func TestPrefix(t *testing.T) {
	aChain := buildAChain()
	expected := []string{"fox"}
	returned := aChain.Prefix("the quick")

	for i, v := range returned {
		if v != expected[i] {
			t.Errorf("The Prefix for 'the quick' was expected to be %s but instead was %s", expected, returned)
		}
	}
}

func TestPrefixes(t *testing.T) {
	aChain := buildAChain()
	expected := []string{" ", " the", "fox jumps", "jumps over", "over the", "quick fox", "the lazy", "the quick"}
	returned := aChain.Prefixes()
	for i, v := range returned {
		if v != expected[i] {
			t.Errorf("The Prefixes in the chain were expected to be [%s] but instead was [%s]", strings.Join(expected, ", "), strings.Join(returned, ", "))
			break
		}
	}
}

func TestBuild(t *testing.T) {
	aChain := buildAChain()
	if l := aChain.Length(); l != 8 {
		t.Errorf("The Chain was erroneously build (length mismatching")
	}
	if s := strings.Join(aChain.Prefix("the quick"), ""); s != "fox" {
		t.Errorf("The Chain was erroneously build (suffix mismatching): %s", s)
	}
	bChain := buildChain(3)
	if s := strings.Join(bChain.Prefix("the quick fox"), ""); s != "jumps" {
		t.Errorf("The Chain was erroneously build (suffix mismatching): %s", s)
	}
}

func TestIsCapital(t *testing.T) {

}

func TestGenerate(t *testing.T) {
	aChain := buildAChain()
	fmt.Println(aChain.Generate(new(bytes.Buffer), 20))
}
