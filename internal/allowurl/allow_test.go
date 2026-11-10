package allowurl

import (
	"errors"
	"testing"
)

func Test_Parse_allows_only_awsapps_device_login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{name: "awsapps device user code", in: "https://revelio.awsapps.com/start/#/device?user_code=MLMD-QGRN", ok: true},
		{name: "awsapps device path without slash", in: "https://revelio.awsapps.com/start#/device?user_code=TEST-TEST", ok: true},
		{name: "awsapps device uppercase host", in: "https://REVELIO.AWSAPPS.COM/start/#/device?user_code=TEST-TEST", ok: true},
		{name: "trims spaces", in: "  https://revelio.awsapps.com/start/#/device?user_code=TEST-TEST  ", ok: true},
		{name: "http rejected", in: "http://revelio.awsapps.com/start/#/device?user_code=TEST-TEST", ok: false},
		{name: "file rejected", in: "file:///etc/passwd", ok: false},
		{name: "javascript rejected", in: "javascript:alert(1)", ok: false},
		{name: "other host rejected", in: "https://evil.example.com/", ok: false},
		{name: "userinfo rejected", in: "https://evil@revelio.awsapps.com/start/#/device?user_code=TEST-TEST", ok: false},
		{name: "awsapps portal rejected", in: "https://revelio.awsapps.com/start/", ok: false},
		{name: "awsapps accounts rejected", in: "https://revelio.awsapps.com/start/#/?tab=accounts", ok: false},
		{name: "awsapps device without code rejected", in: "https://revelio.awsapps.com/start/#/device", ok: false},
		{name: "awsapps other path rejected", in: "https://revelio.awsapps.com/other/#/device?user_code=TEST-TEST", ok: false},
		{name: "awsapps suffix trick rejected", in: "https://revelio.awsapps.com.evil.com/start/#/device?user_code=TEST-TEST", ok: false},
		{name: "empty rejected", in: "   ", ok: false},
		{name: "relative rejected", in: "/login", ok: false},
		{name: "signin authorize", in: "https://us-east-2.signin.aws/platform/authorize?request_uri=urn:ietf:params:oauth:request_uri:abc&identitystore_id=d-9a6706643d", ok: true},
		{name: "signin other path rejected", in: "https://us-east-2.signin.aws/platform/login?request_uri=urn:ietf:params:oauth:request_uri:abc&identitystore_id=d-9a6706643d", ok: false},
		{name: "signin missing request rejected", in: "https://us-east-2.signin.aws/platform/authorize?identitystore_id=d-9a6706643d", ok: false},
		{name: "signin suffix trick rejected", in: "https://us-east-2.signin.aws.evil.com/platform/authorize?request_uri=urn:ietf:params:oauth:request_uri:abc&identitystore_id=d-9a6706643d", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.in)
			if tt.ok {
				if err != nil {
					t.Fatalf("Parse(%q) error = %v", tt.in, err)
				}
				if got.String() == "" {
					t.Fatal("allowed URL has empty string")
				}
				return
			}
			if !errors.Is(err, ErrRejected) {
				t.Fatalf("Parse(%q) error = %v, want ErrRejected", tt.in, err)
			}
		})
	}
}

func Test_Parse_reads_user_code_from_fragment(t *testing.T) {
	t.Parallel()

	got, err := Parse("https://revelio.awsapps.com/start/#/device?user_code=WXYZ-1234")
	if err != nil {
		t.Fatal(err)
	}
	if got.DeviceCode() != "WXYZ-1234" {
		t.Fatalf("fragment code = %q", got.DeviceCode())
	}
}
