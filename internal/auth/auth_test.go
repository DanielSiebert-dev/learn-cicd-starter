package auth

import (
	"errors"
	"net/http"
	"testing"
)

// TestGetAPIKey verifies the behavior of the GetAPIKey function using
// table-driven tests, as recommended by Dave Cheney in his blog post
// "Prefer table driven tests"[](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests).
//
// Key considerations from the post incorporated here:
// - Use a table (slice of structs) to define test cases, reducing duplication
//   and making it easy to add new cases by adding rows.
// - Include a 'name' field in the struct for descriptive test case names,
//   improving readability and debugging.
// - Employ subtests via t.Run to run each case independently, allowing all
//   tests to execute even if one fails, and enabling all
//   tests to execute even if one fails, and enabling targeted runs with
//   'go test -run TestGetAPIKey/subtest-name'.
// - Use an anonymous struct literal to minimize boilerplate.
// - For comparisons, use simple equality for strings; if needed, reflect.DeepEqual
//   could be used for more complex types, with %#v in error messages for better
//   diagnostics (though not necessary here as outputs are primitives).
// - Avoid pitfalls like unnamed tests or stopping on first failure by using
//   subtests instead of t.Fatalf.
// - Note: For complex diffs, consider go-cmp library, but sticking to standard
//   library here for simplicity.
// - Error comparison uses string matching (.Error()) since errors.New creates
//   new instances each time, and errors.Is would fail on pointer inequality.
//
// The test covers valid, missing, and malformed header scenarios to ensure
// robust coverage of boundary conditions. Note that the current function
// implementation allows empty keys ("ApiKey ") and extra parts after the key
// (taking only the second split part), without erroring—tests reflect this
// behavior. If stricter validation is desired, the function could be updated
// to check len(splitAuth) == 2 && splitAuth[1] != "".
func TestGetAPIKey(t *testing.T) {
	// Define the table as a slice of anonymous structs for test cases.
	tests := []struct {
		name    string      // Descriptive name for the test case.
		headers http.Header // Input HTTP headers to simulate the request.
		wantKey string      // Expected API key to be returned.
		wantErr error       // Expected error (nil if no error).
	}{
		{
			name:    "valid authorization header",
			headers: http.Header{"Authorization": []string{"ApiKey my-secret-key"}},
			wantKey: "my-secret-key",
			wantErr: nil,
		},
		{
			name:    "no authorization header",
			headers: http.Header{},
			wantKey: "",
			wantErr: ErrNoAuthHeaderIncluded,
		},
		{
			name:    "malformed header - wrong prefix",
			headers: http.Header{"Authorization": []string{"Bearer some-token"}},
			wantKey: "",
			wantErr: errors.New("malformed authorization header"),
		},
		{
			name:    "empty key after prefix (allowed by function)",
			headers: http.Header{"Authorization": []string{"ApiKey "}},
			wantKey: "",
			wantErr: nil,
		},
		{
			name:    "header with extra parts (takes second part, ignores rest)",
			headers: http.Header{"Authorization": []string{"ApiKey key extra"}},
			wantKey: "key",
			wantErr: nil,
		},
		{
			name:    "case sensitivity in prefix",
			headers: http.Header{"Authorization": []string{"apikey mykey"}},
			wantKey: "",
			wantErr: errors.New("malformed authorization header"),
		},
	}

	// Iterate over the table and run each case as a subtest.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function under test.
			gotKey, gotErr := GetAPIKey(tt.headers)
			// Check the returned key (simple string comparison).
			if gotKey != tt.wantKey {
				t.Errorf("GetAPIKey() gotKey = %#v, want %#v", gotKey, tt.wantKey)
			}
			// Check if error presence matches.
			if (gotErr == nil) != (tt.wantErr == nil) {
				t.Errorf("GetAPIKey() error presence mismatch: got %#v, want %#v", gotErr, tt.wantErr)
			} else if gotErr != nil {
				// Compare error messages as strings for non-constant errors.
				if gotErr.Error() != tt.wantErr.Error() {
					t.Errorf("GetAPIKey() gotErr = %#v, want %#v", gotErr, tt.wantErr)
				}
			}
		})
	}
}