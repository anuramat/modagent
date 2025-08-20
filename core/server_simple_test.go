package core

import (
	"testing"

	"github.com/anuramat/modagent/testutils"
)

// Simple tests that don't require command mocking
func TestParseArgsSimple(t *testing.T) {
	tests := []testutils.TableTest{
		{
			Name: "valid basic args",
			Input: map[string]any{
				"prompt": "test prompt",
			},
			Expected: CallArgs{
				Prompt: " test prompt",
			},
			WantErr: false,
		},
		{
			Name:     "missing prompt",
			Input:    map[string]any{},
			Expected: CallArgs{},
			WantErr:  true,
		},
	}

	testutils.RunTableTests(t, tests, func(t *testing.T, tt testutils.TableTest) {
		result, err := ParseArgs(tt.Input.(map[string]any))

		if tt.WantErr {
			testutils.AssertError(t, err)
			return
		}

		testutils.AssertNoError(t, err)
		expected := tt.Expected.(CallArgs)
		testutils.AssertEqual(t, expected.Prompt, result.Prompt)
	})
}

func TestBuildResponseSimple(t *testing.T) {
	tests := []testutils.TableTest{
		{
			Name: "basic response",
			Input: struct {
				output         string
				conversationID string
				tempDir        string
				jsonOutput     bool
			}{"test output", "conv123", "", false},
			Expected: `{"conversation":"conv123","response":"test output"}`,
		},
	}

	testutils.RunTableTests(t, tests, func(t *testing.T, tt testutils.TableTest) {
		input := tt.Input.(struct {
			output         string
			conversationID string
			tempDir        string
			jsonOutput     bool
		})

		result, err := buildResponse(input.output, input.conversationID, input.tempDir, input.jsonOutput)

		testutils.AssertNoError(t, err)
		testutils.AssertJSONEqual(t, tt.Expected.(string), result)
	})
}

func TestExtractConversationIDSimple(t *testing.T) {
	tests := []testutils.TableTest{
		{
			Name:     "valid conversation line",
			Input:    "Some output\nConversation saved: abc123\n",
			Expected: "abc123",
		},
		{
			Name:     "no conversation line",
			Input:    "Some output\nNo conversation here",
			Expected: "",
		},
		{
			Name:     "empty input",
			Input:    "",
			Expected: "",
		},
	}

	testutils.RunTableTests(t, tests, func(t *testing.T, tt testutils.TableTest) {
		result := extractConversationID(tt.Input.(string))
		testutils.AssertEqual(t, tt.Expected, result)
	})
}
