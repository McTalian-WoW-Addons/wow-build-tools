package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisableEmoji(t *testing.T) {
	// Store original values
	origBuild := Build
	origCurseForge := CurseForge
	origDirectory := Directory
	origDone := Done
	origExternal := External
	origFile := File
	origFinish := Finish
	origGitHub := GitHub
	origIgnore := Ignore
	origInject := Inject
	origPackage := Package
	origProcessing := Processing
	origTime := Time
	origWago := Wago
	origWatch := Watch
	origWoWInterface := WoWInterface
	origZip := Zip
	origZipFile := ZipFile

	// Ensure originals are not empty
	assert.NotEmpty(t, origBuild)
	assert.NotEmpty(t, origCurseForge)

	// Disable emojis
	DisableEmoji()

	// Check all are empty
	assert.Empty(t, Build)
	assert.Empty(t, CurseForge)
	assert.Empty(t, Directory)
	assert.Empty(t, Done)
	assert.Empty(t, External)
	assert.Empty(t, File)
	assert.Empty(t, Finish)
	assert.Empty(t, GitHub)
	assert.Empty(t, Ignore)
	assert.Empty(t, Inject)
	assert.Empty(t, Package)
	assert.Empty(t, Processing)
	assert.Empty(t, Time)
	assert.Empty(t, Wago)
	assert.Empty(t, Watch)
	assert.Empty(t, WoWInterface)
	assert.Empty(t, Zip)
	assert.Empty(t, ZipFile)

	// Restore original values for other tests
	Build = origBuild
	CurseForge = origCurseForge
	Directory = origDirectory
	Done = origDone
	External = origExternal
	File = origFile
	Finish = origFinish
	GitHub = origGitHub
	Ignore = origIgnore
	Inject = origInject
	Package = origPackage
	Processing = origProcessing
	Time = origTime
	Wago = origWago
	Watch = origWatch
	WoWInterface = origWoWInterface
	Zip = origZip
	ZipFile = origZipFile
}

func TestEnableEmoji(t *testing.T) {
	// Disable emojis first
	DisableEmoji()

	// Verify they are empty
	assert.Empty(t, Build)
	assert.Empty(t, CurseForge)
	assert.Empty(t, Directory)

	// Re-enable emojis
	EnableEmoji()

	// Check all are not empty
	assert.NotEmpty(t, Build)
	assert.NotEmpty(t, CurseForge)
	assert.NotEmpty(t, Directory)
	assert.NotEmpty(t, Done)
	assert.NotEmpty(t, External)
	assert.NotEmpty(t, File)
	assert.NotEmpty(t, Finish)
	assert.NotEmpty(t, GitHub)
	assert.NotEmpty(t, Ignore)
	assert.NotEmpty(t, Inject)
	assert.NotEmpty(t, Package)
	assert.NotEmpty(t, Processing)
	assert.NotEmpty(t, Time)
	assert.NotEmpty(t, Wago)
	assert.NotEmpty(t, Watch)
	assert.NotEmpty(t, WoWInterface)
	assert.NotEmpty(t, Zip)
	assert.NotEmpty(t, ZipFile)

	// Verify specific emojis
	assert.Equal(t, "🔨 ", Build)
	assert.Equal(t, "🔥 ", CurseForge)
	assert.Equal(t, "📦 ", ZipFile)
}

func TestDisableAndEnableEmoji(t *testing.T) {
	// Store originals
	origBuild := Build

	// Disable
	DisableEmoji()
	assert.Empty(t, Build)

	// Enable
	EnableEmoji()
	assert.NotEmpty(t, Build)
	assert.Equal(t, origBuild, Build)
}
