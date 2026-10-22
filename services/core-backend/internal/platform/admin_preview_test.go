package platform

import "testing"

func TestAdminPreviewRestrictedToLocalCopy(t *testing.T) {
	for _, tc := range []struct {
		name, env, bind, db, flag string
		allowed                   bool
	}{
		{"local copy", "development", "127.0.0.1", "postgres://preview@127.0.0.1:55440/hackways_admin_preview", "true", true},
		{"production mode", "production", "127.0.0.1", "postgres://preview@127.0.0.1:55440/hackways_admin_preview", "true", false},
		{"public listener", "development", "0.0.0.0", "postgres://preview@127.0.0.1:55440/hackways_admin_preview", "true", false},
		{"remote database", "development", "127.0.0.1", "postgres://preview@database.example.invalid/hackways_admin_preview", "true", false},
		{"production database name", "development", "127.0.0.1", "postgres://preview@127.0.0.1/hackways_prod", "true", false},
		{"query host override", "development", "127.0.0.1", "postgres://preview@127.0.0.1:55440/hackways_admin_preview?host=database.example.invalid", "true", false},
		{"query database override", "development", "127.0.0.1", "postgres://preview@127.0.0.1:55440/hackways_admin_preview?dbname=hackways_prod", "true", false},
		{"malformed flag", "development", "127.0.0.1", "postgres://preview@127.0.0.1/hackways_admin_preview", "perhaps", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			enabled, err := adminPreviewMode(tc.env, tc.bind, tc.db, tc.flag)
			if tc.allowed {
				if err != nil || !enabled {
					t.Fatal("valid local preview was refused")
				}
			} else if err == nil {
				t.Fatal("unsafe preview configuration accepted")
			}
		})
	}
	if enabled, err := adminPreviewMode("production", "0.0.0.0", "postgres://database/hackways_prod", ""); err != nil || enabled {
		t.Fatal("ordinary runtime behavior changed")
	}
}
