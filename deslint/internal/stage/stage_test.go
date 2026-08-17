package stage_test

import (
	"os"
	"testing"

	"github.com/aprilgom/AnslDes/deslint/internal/stage"
)

func TestParsePreservesCommandOutputAndFreshness(t *testing.T) {
	t.Parallel()
	contents, err := os.ReadFile("../../../packages/schema/testdata/stage-execution-positive.json")
	if err != nil {
		t.Fatal(err)
	}
	value, err := stage.Parse(contents)
	if err != nil || !stage.Fresh(value) || value.Command[0] != "example-provider" || value.Stdout == "" {
		t.Fatalf("Parse() = %#v, %v", value, err)
	}
}
