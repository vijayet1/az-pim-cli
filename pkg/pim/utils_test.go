/*
Copyright © 2024 netr0m <netr0m@pm.me>
*/
package pim

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDateTime(t *testing.T) {
	now := time.Now().Local()
	currentDate := now.Format("2006-01-02")
	currentTZ := now.Format("-07:00")
	errMsg := "resulting startDateTime does not match expected value"

	dateOnly, _ := parseDateTime("31/12/2024", "")
	timeOnly, _ := parseDateTime("", "13:37")
	dateTime, _ := parseDateTime("31/12/2024", "13:37")

	assert.Equal(t, fmt.Sprintf("2024-12-31T00:00:00%s", currentTZ), dateOnly, errMsg)
	assert.Equal(t, fmt.Sprintf("%sT13:37:00%s", currentDate, currentTZ), timeOnly, errMsg)
	assert.Equal(t, fmt.Sprintf("2024-12-31T13:37:00%s", currentTZ), dateTime, errMsg)
}

func TestCreateResourceAssignmentRequest(t *testing.T) {
	resourceAssignment := &EligibleResourceAssignmentsDummyData.Value[0]
	tests := []struct {
		name     string
		scope    string
		expected string
	}{
		{
			name:     "without resource scope",
			scope:    "",
			expected: resourceAssignment.Properties.ExpandedProperties.Scope.Id,
		},
		{
			name:     "with resource scope",
			scope:    fmt.Sprintf("/subscriptions/%s/resourceGroups/rg", TEST_DUMMY_SUBSCRIPTION_1_ID),
			expected: fmt.Sprintf("/subscriptions/%s/resourceGroups/rg", TEST_DUMMY_SUBSCRIPTION_1_ID),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, _ := CreateResourceAssignmentRequestWithScope(TEST_DUMMY_PRINCIPAL_ID, resourceAssignment, tt.scope, 30, "", "", "test", "Test", "1337")

			assert.Equal(t, tt.expected, scope)
		})
	}
}
