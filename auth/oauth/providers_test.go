package oauth

import (
	"aigateway/auth/pkce"
	"strings"
	"testing"
)

func TestGetProviderOAuth(t *testing.T) {
	tests := []struct {
		name        string
		providerID  string
		redirectURI string
		wantErr     bool
	}{
		{"antigravity", "antigravity", "http://localhost:8088/callback", false},
		{"codex", "codex", "http://localhost:8088/callback", false},
		{"claude", "claude", "http://localhost:8088/callback", false},
		{"unknown", "unknown", "http://localhost:8088/callback", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := GetProviderOAuth(tt.providerID, tt.redirectURI)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProviderOAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("GetProviderOAuth() returned nil provider")
			}
			if !tt.wantErr && provider.ProviderID != tt.providerID {
				t.Errorf("ProviderID = %s, want %s", provider.ProviderID, tt.providerID)
			}
		})
	}
}

func TestBuildAuthURL(t *testing.T) {
	redirectURI := "http://localhost:8088/callback"
	state := "test-state-123"

	pkceCodes, err := pkce.GeneratePKCECodes()
	if err != nil {
		t.Fatalf("Failed to generate PKCE codes: %v", err)
	}

	tests := []struct {
		name          string
		providerID    string
		wantContains  []string
		wantNotContain []string
	}{
		{
			name:       "antigravity",
			providerID: "antigravity",
			wantContains: []string{
				"accounts.google.com",
				"client_id=",
				"redirect_uri=",
				"state=" + state,
				"code_challenge=",
				"code_challenge_method=S256",
				"access_type=offline",
			},
		},
		{
			name:       "codex",
			providerID: "codex",
			wantContains: []string{
				"auth.openai.com",
				"codex_cli_simplified_flow=true",
				"id_token_add_organizations=true",
			},
		},
		{
			name:       "claude",
			providerID: "claude",
			wantContains: []string{
				"claude.ai",
				"code=true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, _ := GetProviderOAuth(tt.providerID, redirectURI)
			authURL, err := provider.BuildAuthURL(state, pkceCodes)
			if err != nil {
				t.Fatalf("BuildAuthURL() error = %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(authURL, want) {
					t.Errorf("AuthURL missing %q, got %s", want, authURL)
				}
			}
		})
	}
}

func TestBuildAuthURLRequiresPKCE(t *testing.T) {
	provider, _ := GetProviderOAuth("antigravity", "http://localhost:8088/callback")

	_, err := provider.BuildAuthURL("test-state", nil)
	if err == nil {
		t.Error("BuildAuthURL() should error when PKCE codes are nil")
	}
}

func TestListProviders(t *testing.T) {
	providers := ListProviders("http://localhost:8088/callback")

	if len(providers) != 3 {
		t.Errorf("ListProviders() returned %d providers, want 3", len(providers))
	}

	providerIDs := make(map[string]bool)
	for _, p := range providers {
		providerIDs[p.ProviderID] = true
	}

	expectedIDs := []string{"antigravity", "codex", "claude"}
	for _, id := range expectedIDs {
		if !providerIDs[id] {
			t.Errorf("Missing provider %s in ListProviders()", id)
		}
	}
}

func TestProviderOAuthFields(t *testing.T) {
	tests := []struct {
		providerID   string
		wantName     string
		wantAuthURL  string
		wantTokenURL string
	}{
		{
			providerID:   "antigravity",
			wantName:     "Google Cloud Code (Antigravity)",
			wantAuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			wantTokenURL: "https://oauth2.googleapis.com/token",
		},
		{
			providerID:   "codex",
			wantName:     "OpenAI Codex",
			wantAuthURL:  "https://auth.openai.com/oauth/authorize",
			wantTokenURL: "https://auth.openai.com/oauth/token",
		},
		{
			providerID:   "claude",
			wantName:     "Anthropic Claude",
			wantAuthURL:  "https://claude.ai/oauth/authorize",
			wantTokenURL: "https://console.anthropic.com/v1/oauth/token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.providerID, func(t *testing.T) {
			provider, _ := GetProviderOAuth(tt.providerID, "http://localhost/callback")

			if provider.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", provider.Name, tt.wantName)
			}
			if provider.AuthURL != tt.wantAuthURL {
				t.Errorf("AuthURL = %s, want %s", provider.AuthURL, tt.wantAuthURL)
			}
			if provider.TokenURL != tt.wantTokenURL {
				t.Errorf("TokenURL = %s, want %s", provider.TokenURL, tt.wantTokenURL)
			}
		})
	}
}
