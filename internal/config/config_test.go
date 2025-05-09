package config

import (
	"testing"
)

func TestIsValidPostgresIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantValid   bool
		wantReason  string
	}{
		{
			name:      "valid unquoted name",
			input:     "valid_name",
			wantValid: true,
			wantReason: "",
		},
		{
			name:      "valid mixed case with numbers",
			input:     "Valid_Name_123",
			wantValid: true,
			wantReason: "",
		},
		{
			name:      "valid starting with underscore",
			input:     "_valid_name",
			wantValid: true,
			wantReason: "",
		},
		{
			name:      "valid quoted name with spaces",
			input:     `"Quoted Name"`,
			wantValid: true,
			wantReason: "",
		},
		{
			name:      "valid quoted name with escaped quotes",
			input:     `"Quoted""Name"`,
			wantValid: true,
			wantReason: "",
		},
		{
			name:      "invalid starting with number",
			input:     "123invalid",
			wantValid: false,
			wantReason: "identifier must begin with a letter or underscore",
		},
		{
			name:      "invalid contains hyphen",
			input:     "invalid-name",
			wantValid: false,
			wantReason: "identifier contains invalid character: -",
		},
		{
			name:      "invalid reserved keyword",
			input:     "select",
			wantValid: false,
			wantReason: "select is a reserved keyword",
		},
		{
			name:      "invalid too long",
			input:     "verylongidentifierthatexceedsthirtychars",
			wantValid: false,
			wantReason: "identifier too long (maximum is 31 characters)",
		},
		{
			name:      "invalid empty quoted identifier",
			input:     `""`,
			wantValid: false,
			wantReason: "quoted identifier cannot be empty",
		},
		{
			name:      "invalid empty identifier",
			input:     "",
			wantValid: false,
			wantReason: "identifier cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, gotReason := isValidPostgresIdentifier(tt.input)
			if gotValid != tt.wantValid {
				t.Errorf("IsValidPostgresIdentifier() valid = %v, want %v", gotValid, tt.wantValid)
			}
			if gotReason != tt.wantReason {
				t.Errorf("IsValidPostgresIdentifier() reason = %v, want %v", gotReason, tt.wantReason)
			}
		})
	}
}