package memory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ahmad-ubaidillah/aigo/pkg/types"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) *SessionDB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := NewSessionDB(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestProfile_CreateAndGet(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	profile, err := db.CreateProfile("p1", "default")
	require.NoError(t, err)
	require.Equal(t, "p1", profile.ID)
	require.Equal(t, "default", profile.Name)
	require.NotNil(t, profile.PlatformPrefs)
	require.NotNil(t, profile.Preferences)

	got, err := db.GetProfile("p1")
	require.NoError(t, err)
	require.Equal(t, "p1", got.ID)
	require.Equal(t, "default", got.Name)
}

func TestProfile_GetNotFound(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.GetProfile("nonexistent")
	require.Error(t, err)
}

func TestProfile_GetByName(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.CreateProfile("p1", "myprofile")
	require.NoError(t, err)

	got, err := db.GetProfileByName("myprofile")
	require.NoError(t, err)
	require.Equal(t, "p1", got.ID)
	require.Equal(t, "myprofile", got.Name)
}

func TestProfile_GetByName_NotFound(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.GetProfileByName("nonexistent")
	require.Error(t, err)
}

func TestProfile_List(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.CreateProfile("p1", "alpha")
	require.NoError(t, err)
	_, err = db.CreateProfile("p2", "beta")
	require.NoError(t, err)

	profiles, err := db.ListProfiles()
	require.NoError(t, err)
	require.Len(t, profiles, 2)
}

func TestProfile_Update(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	profile, err := db.CreateProfile("p1", "default")
	require.NoError(t, err)

	profile.DefaultModel = "gpt-4"
	profile.CodingModel = "claude-3"
	profile.OpenCodeTimeout = 60
	profile.OpenCodeMaxTurns = 20
	profile.Preferences["theme"] = "dark"
	profile.PlatformPrefs["telegram"] = "token123"

	err = db.UpdateProfile(profile)
	require.NoError(t, err)

	got, err := db.GetProfile("p1")
	require.NoError(t, err)
	require.Equal(t, "gpt-4", got.DefaultModel)
	require.Equal(t, "claude-3", got.CodingModel)
	require.Equal(t, 60, got.OpenCodeTimeout)
	require.Equal(t, 20, got.OpenCodeMaxTurns)
	require.Equal(t, "dark", got.Preferences["theme"])
	require.Equal(t, "token123", got.PlatformPrefs["telegram"])
}

func TestProfile_Delete(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.CreateProfile("p1", "default")
	require.NoError(t, err)

	err = db.DeleteProfile("p1")
	require.NoError(t, err)

	_, err = db.GetProfile("p1")
	require.Error(t, err)
}

func TestProfile_SetPreference(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.CreateProfile("p1", "default")
	require.NoError(t, err)

	err = db.SetProfilePreference("p1", "theme", "dark")
	require.NoError(t, err)

	got, err := db.GetProfile("p1")
	require.NoError(t, err)
	require.Equal(t, "dark", got.Preferences["theme"])
}

func TestProfile_SetPlatformPreference(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.CreateProfile("p1", "default")
	require.NoError(t, err)

	err = db.SetPlatformPreference("p1", "telegram", "tok123")
	require.NoError(t, err)

	got, err := db.GetProfile("p1")
	require.NoError(t, err)
	require.Equal(t, "tok123", got.PlatformPrefs["telegram"])
}

func TestProfile_DuplicateName(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.CreateProfile("p1", "default")
	require.NoError(t, err)

	_, err = db.CreateProfile("p2", "default")
	require.Error(t, err)
}

func TestProfile_CreateProfile_FileExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := NewSessionDB(dbPath)
	require.NoError(t, err)

	_, err = db.CreateProfile("p1", "first")
	require.NoError(t, err)
	db.Close()

	db2, err := NewSessionDB(dbPath)
	require.NoError(t, err)
	defer db2.Close()

	got, err := db2.GetProfile("p1")
	require.NoError(t, err)
	require.Equal(t, "first", got.Name)
}

func TestProfile_SetPreference_NonExistent(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	err := db.SetProfilePreference("nonexistent", "key", "val")
	require.Error(t, err)
}

func TestProfile_SetPlatformPreference_NonExistent(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	err := db.SetPlatformPreference("nonexistent", "telegram", "tok")
	require.Error(t, err)
}

func TestProfile_Update_NonExistent(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	err := db.UpdateProfile(&types.Profile{ID: "nonexistent"})
	require.Error(t, err)
}

func TestProfile_Delete_NonExistent(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	err := db.DeleteProfile("nonexistent")
	require.Error(t, err)
}

func TestProfile_MultipleProfiles(t *testing.T) {
	t.Parallel()

	db := newTestDB(t)

	_, err := db.CreateProfile("p1", "work")
	require.NoError(t, err)
	_, err = db.CreateProfile("p2", "personal")
	require.NoError(t, err)
	_, err = db.CreateProfile("p3", "testing")
	require.NoError(t, err)

	profiles, err := db.ListProfiles()
	require.NoError(t, err)
	require.Len(t, profiles, 3)

	names := make(map[string]bool)
	for _, p := range profiles {
		names[p.Name] = true
	}
	require.True(t, names["work"])
	require.True(t, names["personal"])
	require.True(t, names["testing"])
}

func TestProfile_CleanupOnClose(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := NewSessionDB(dbPath)
	require.NoError(t, err)
	_, err = db.CreateProfile("p1", "default")
	require.NoError(t, err)
	db.Close()

	_, err = os.Stat(dbPath)
	require.NoError(t, err)
}
