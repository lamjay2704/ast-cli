//go:build integration

package integration

import (
	"fmt"
	"io"
	"testing"

	"github.com/checkmarx/ast-cli/internal/commands/util/printer"
	"github.com/checkmarx/ast-cli/internal/params"
	"github.com/checkmarx/ast-cli/internal/wrappers"
	"gotest.tools/assert"
)

// - SAST : Update Severity, Status and Comment for a given similarity id
// - SAST : Get all predicates for a given project and similarity id

var projectID string

func TestSastUpdateAndGetPredicatesForSimilarityId(t *testing.T) {

	fmt.Println("Step 1: Testing the command 'triage update' to update an issue from the project.")

	_, projectID = getRootScan(t)
	similarityID := "1826563305"
	state := "Urgent"
	severity := "Medium"
	comment := "Testing CLI Command for triage."
	scanType := "sast"

	outputBufferForStep1 := executeCmdNilAssertion(
		t, "Issue should be updated.", "triage", "update",
		flag(params.ProjectIDFlag), projectID,
		flag(params.SimilarityIDFlag), similarityID,
		flag(params.StateFlag), state,
		flag(params.SeverityFlag), severity,
		flag(params.CommentFlag), comment,
		flag(params.ScanTypeFlag), scanType,
	)

	_, readingError := io.ReadAll(outputBufferForStep1)
	assert.NilError(t, readingError, "Reading result should pass")

	fmt.Println("Step 2: Testing the command 'triage show' to get the same predicate back.")
	outputBufferForStep2 := executeCmdNilAssertion(
		t, "Predicates should be fetched.", "triage", "show",
		flag(params.FormatFlag), printer.FormatJSON,
		flag(params.ProjectIDFlag), projectID,
		flag(params.SimilarityIDFlag), similarityID,
		flag(params.ScanTypeFlag), scanType,
	)

	result := []wrappers.Predicate{}
	fmt.Println(outputBufferForStep2)
	_ = unmarshall(t, outputBufferForStep2, &result, "Reading results should pass")

	assert.Assert(t, (len(result)) == 1, "Should have 1 predicate as the result.")

	deleteScanAndProject()

}

func TestSastUpdatePredicateWithInvalidValues(t *testing.T) {

	err, _ := executeCommand(
		t, "triage", "update",
		flag(params.ProjectIDFlag), "1234",
		flag(params.SimilarityIDFlag), "1234",
		flag(params.StateFlag), "URGENT",
		flag(params.SeverityFlag), "High",
		flag(params.CommentFlag), "",
		flag(params.ScanTypeFlag), "sast",
	)

	fmt.Println(err)
	assertError(t, err, "Predicate bad request")
}
