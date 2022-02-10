//go:build integration

package integration

import (
	"testing"

	"github.com/checkmarx/ast-cli/internal/params"
	"github.com/checkmarx/ast-cli/internal/wrappers"
)

func TestRunGetBflByScanIdAndQueryId(t *testing.T) {

	assertRequiredParameter(t, "required flag(s) \"query-id\", \"scan-id\" not set", "results", "bfl")
	scanID, _ := getRootScan(t)
	queryId := "17765437696070740537"

	outputBufferForStep2 := executeCmdNilAssertion(
		t, "Getting BFL should pass.", "results", "bfl",
		flag(params.ScanIDFlag), scanID,
		flag(params.QueryIdFlag), queryId,
		flag(params.FormatFlag), "json")

	bflResult := []wrappers.ScanResultNode{}
	_ = unmarshall(t, outputBufferForStep2, &bflResult, "Reading BFL results should pass")

}
