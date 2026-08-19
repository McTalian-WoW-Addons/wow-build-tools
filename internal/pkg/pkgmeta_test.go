package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/McTalian/wow-build-tools/internal/external"
)

func TestPkgMeta_UnmarshalYAML(t *testing.T) {
	yamlData := `
package-as: test-package
enable-nolib-creation: true
required-dependencies:
  - dep1
  - dep2
ignore:
  - ignore1
  - ignore2
move-folders:
  src1: dest1
  src2: dest2
externals:
  ext1:
    type: git
    url: https://example.com/repo.git
  ext2:
    type: svn
    url: https://example.com/repo2.svn

wowi-archive-previous: false
`

	pkgMeta := defaultPkgMeta()
	err := yaml.Unmarshal([]byte(yamlData), &pkgMeta)
	require.NoError(t, err, "Unmarshal failed")

	assert.Equal(t, "test-package", pkgMeta.PackageAs, "PackageAs mismatch")
	assert.True(t, pkgMeta.EnableNoLibCreation, "Expected EnableNoLibCreation to be true")
	assert.Len(t, pkgMeta.RequiredDependencies, 2, "Expected 2 RequiredDependencies")
	assert.Equal(t, "dep1", pkgMeta.RequiredDependencies[0], "RequiredDependencies[0] mismatch")
	assert.Equal(t, "dep2", pkgMeta.RequiredDependencies[1], "RequiredDependencies[1] mismatch")
	assert.Len(t, pkgMeta.Ignore, 2, "Expected 2 Ignore")
	assert.Equal(t, "ignore1", pkgMeta.Ignore[0], "Ignore[0] mismatch")
	assert.Equal(t, "ignore2", pkgMeta.Ignore[1], "Ignore[1] mismatch")
	assert.Len(t, pkgMeta.MoveFolders, 2, "Expected 2 MoveFolders")
	assert.Equal(t, "dest1", pkgMeta.MoveFolders["src1"], "MoveFolders[src1] mismatch")
	assert.Equal(t, "dest2", pkgMeta.MoveFolders["src2"], "MoveFolders[src2] mismatch")
	assert.Len(t, pkgMeta.Externals, 2, "Expected 2 Externals")
	assert.True(t, pkgMeta.ManualChangelog.MarkupType == "text", "Expected ManualChangelog.MarkupType to be 'text'")
	assert.False(t, pkgMeta.WowiArchivePrevious, "Expected WowiArchivePrevious to be false")
	assert.True(t, pkgMeta.WowiConvertChangelog, "Expected WowiConvertChangelog to be true")
	assert.True(t, pkgMeta.WowiCreateChangelog, "Expected WowiCreateChangelog to be true")
}

// TestFetchExternals_SvnConstructorErrorDoesNotHang locks in a fix for a bug
// where checkoutWg.Add(1) ran before NewSvnExternal, so an early "continue"
// on constructor failure (e.g. svn missing from PATH) left the WaitGroup
// counter permanently incremented with no matching Done -- checkoutWg.Wait()
// then blocked forever instead of FetchExternals returning the error.
func TestFetchExternals_SvnConstructorErrorDoesNotHang(t *testing.T) {
	t.Setenv("PATH", t.TempDir()) // hide svn so external.NewSvnExternal errors

	pkgMeta := &PkgMeta{
		Externals: map[string]*external.ExternalEntry{
			"Libs/Foo": {
				EType:    external.Svn,
				URL:      "https://example.com/svn/foo/trunk",
				DestPath: "Libs/Foo",
			},
		},
	}

	done := make(chan error, 1)
	go func() {
		done <- pkgMeta.FetchExternals(t.TempDir(), false, 24*time.Hour)
	}()

	select {
	case err := <-done:
		assert.Error(t, err, "expected the missing-svn constructor error to surface")
	case <-time.After(5 * time.Second):
		t.Fatal("FetchExternals did not return -- checkoutWg.Wait() likely hung on an unmatched Add/Done")
	}
}
