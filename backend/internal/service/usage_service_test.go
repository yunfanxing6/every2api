package service

import "testing"

func TestSelectBalanceEmailNotification(t *testing.T) {
	tests := []struct {
		name     string
		previous float64
		current  float64
		want     balanceEmailNotificationType
	}{
		{name: "no threshold crossed", previous: 8, current: 6, want: balanceEmailNotificationNone},
		{name: "cross low balance threshold", previous: 5, current: 4.99, want: balanceEmailNotificationLow},
		{name: "already below threshold", previous: 4.5, current: 3.5, want: balanceEmailNotificationNone},
		{name: "cross exhausted threshold", previous: 1.2, current: 0, want: balanceEmailNotificationExhausted},
		{name: "jump to negative uses exhausted notice only", previous: 8, current: -1, want: balanceEmailNotificationExhausted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectBalanceEmailNotification(tt.previous, tt.current)
			if got != tt.want {
				t.Fatalf("selectBalanceEmailNotification() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeBalanceEmailBranding(t *testing.T) {
	tests := []struct {
		name        string
		siteName    string
		frontendURL string
		wantName    string
		wantURL     string
	}{
		{
			name:        "fallback to custom deployment defaults",
			siteName:    "",
			frontendURL: "",
			wantName:    defaultBalanceEmailSiteName,
			wantURL:     defaultBalanceEmailFrontendURL,
		},
		{
			name:        "replace stock site name",
			siteName:    "Sub2API",
			frontendURL: "https://example.com/",
			wantName:    defaultBalanceEmailSiteName,
			wantURL:     "https://example.com",
		},
		{
			name:        "preserve custom branding",
			siteName:    "My Relay",
			frontendURL: "https://relay.example.com/app/",
			wantName:    "My Relay",
			wantURL:     "https://relay.example.com/app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotURL := normalizeBalanceEmailBranding(tt.siteName, tt.frontendURL)
			if gotName != tt.wantName || gotURL != tt.wantURL {
				t.Fatalf("normalizeBalanceEmailBranding() = (%q, %q), want (%q, %q)", gotName, gotURL, tt.wantName, tt.wantURL)
			}
		})
	}
}
