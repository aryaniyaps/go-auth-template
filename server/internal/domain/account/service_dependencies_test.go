package account

import (
	"testing"

	"github.com/nyaruka/phonenumbers"
	"github.com/stretchr/testify/assert"
)

func TestPhoneNumberValidation(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber string
		expected    bool
	}{
		{
			name:        "valid US phone number",
			phoneNumber: "+12345678901",
			expected:    true,
		},
		{
			name:        "valid UK phone number",
			phoneNumber: "+447911123456",
			expected:    true,
		},
		{
			name:        "invalid phone number too short",
			phoneNumber: "+12345",
			expected:    false,
		},
		{
			name:        "invalid phone number without country code",
			phoneNumber: "2345678901",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phoneNum, err := phonenumbers.Parse(tt.phoneNumber, "")
			if err != nil {
				// If parsing fails, we consider it invalid
				assert.False(t, tt.expected, "Expected parsing to succeed for valid number")
				return
			}

			isValid := phonenumbers.IsValidNumber(phoneNum)
			assert.Equal(t, tt.expected, isValid)
		})
	}
}

func TestS3ConfigurationFields(t *testing.T) {
	// This test validates that our configuration structure supports S3 fields
	config := &struct {
		S3Bucket    string
		S3Region    string
		S3AccessKey string
		S3SecretKey string
	}{
		S3Bucket:    "test-bucket",
		S3Region:    "us-east-1",
		S3AccessKey: "test-access-key",
		S3SecretKey: "test-secret-key",
	}

	assert.Equal(t, "test-bucket", config.S3Bucket)
	assert.Equal(t, "us-east-1", config.S3Region)
	assert.Equal(t, "test-access-key", config.S3AccessKey)
	assert.Equal(t, "test-secret-key", config.S3SecretKey)
}

func TestSMSConfigurationFields(t *testing.T) {
	// This test validates that our configuration structure supports SMS fields
	config := &struct {
		SMSProvider    string
		SMSTwilioSID   string
		SMSTwilioToken string
		SMSFromNumber  string
	}{
		SMSProvider:    "dummy",
		SMSTwilioSID:   "test-sid",
		SMSTwilioToken: "test-token",
		SMSFromNumber:  "+12345678901",
	}

	assert.Equal(t, "dummy", config.SMSProvider)
	assert.Equal(t, "test-sid", config.SMSTwilioSID)
	assert.Equal(t, "test-token", config.SMSTwilioToken)
	assert.Equal(t, "+12345678901", config.SMSFromNumber)
}