package staticlint

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestOsExitInMain(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer)
}
