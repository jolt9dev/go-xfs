package xfs_test

import (
	"io/fs"
	"os"
	"runtime"
	"testing"

	"github.com/jolt9dev/go-xfs"
	"github.com/stretchr/testify/assert"
)

func init() {
	xfs.MkdirAllDefault("testdir")
	xfs.WriteFile("testfile", []byte("test data"), 0644)
	xfs.Symlink("testfile", "testsymlink")
}

func TestChown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Chown is not supported on Windows")
	}

	err := xfs.Chown("testfile", 1000, 1000)
	assert.NoError(t, err)
}

func TestChmod(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Chmod is not supported on Windows")
	}

	err := xfs.Chmod("testfile", 0644)
	assert.NoError(t, err)
}

func TestCopy(t *testing.T) {
	defer xfs.Remove("testfile_copy")

	err := xfs.Copy("testfile", "testfile_copy", true)
	assert.NoError(t, err)
}

func TestCopyDir(t *testing.T) {
	defer xfs.RemoveAll("testdir_copy")

	err := xfs.CopyDir("testdir", "testdir_copy", true)
	assert.NoError(t, err)
}

func TestCopyFile(t *testing.T) {
	defer xfs.Remove("testfile_copy")

	err := xfs.CopyFile("testfile", "testfile_copy", true)
	assert.NoError(t, err)
}

func TestCreate(t *testing.T) {
	defer xfs.Remove("testfile2")
	file, err := xfs.Create("testfile2")
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()
}

func TestCreateTemp(t *testing.T) {
	file, err := xfs.CreateTemp("", "testfile")
	if err == nil {
		defer xfs.Remove(file.Name())
	}
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()
}

func TestCwd(t *testing.T) {
	dir, err := xfs.Cwd()
	assert.NoError(t, err)
	assert.NotEmpty(t, dir)
}

func TestExists(t *testing.T) {
	exists := xfs.Exists("testfile")
	assert.True(t, exists)

	exists = xfs.Exists("testdir")
	assert.True(t, exists)

	exists = xfs.Exists("testfile999")
	assert.False(t, exists)
}

func TestEnsureDir(t *testing.T) {
	err := xfs.EnsureDir("testdir", 0755)
	assert.NoError(t, err)

	defer xfs.RemoveAll("testdir10")
	err = xfs.EnsureDir("testdir10", 0755)
	assert.NoError(t, err)
}

func TestEnsureDirDefault(t *testing.T) {
	err := xfs.EnsureDirDefault("testdir")
	assert.NoError(t, err)
}

func TestEnsureFile(t *testing.T) {
	err := xfs.EnsureFile("testfile", 0644)
	assert.NoError(t, err)

	defer xfs.Remove("testfile10")
	err = xfs.EnsureFile("testfile10", 0644)
	assert.NoError(t, err)
}

func TestEnsureFileDefault(t *testing.T) {
	err := xfs.EnsureFileDefault("testfile")
	assert.NoError(t, err)
}

func TestIsFile(t *testing.T) {
	isFile := xfs.IsFile("testfile")
	assert.True(t, isFile)
}

func TestIsDir(t *testing.T) {
	isDir := xfs.IsDir("testdir")
	assert.True(t, isDir)
}

func TestIsSymlink(t *testing.T) {
	defer xfs.Remove("testsymlink")
	isSymlink := xfs.IsSymlink("testsymlink")
	assert.True(t, isSymlink)
}

func TestLink(t *testing.T) {
	defer xfs.Remove("testfile_link")
	err := xfs.Link("testfile", "testfile_link")
	assert.NoError(t, err)
}

func TestLstat(t *testing.T) {
	info, err := xfs.Lstat("testfile")
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestMkdir(t *testing.T) {
	defer xfs.Remove("testdir700")
	err := xfs.Mkdir("testdir700", 0755)
	assert.NoError(t, err)
}

func TestMkdirDefault(t *testing.T) {
	defer xfs.Remove("testdir800")
	err := xfs.MkdirDefault("testdir800")
	assert.NoError(t, err)
}

func TestMkdirAll(t *testing.T) {
	defer xfs.RemoveAll("testdir900")
	err := xfs.MkdirAll("testdir900/subdir", 0755)
	assert.NoError(t, err)
}

func TestMkdirAllDefault(t *testing.T) {
	defer xfs.RemoveAll("testdir2000")
	err := xfs.MkdirAllDefault("testdir2000/subdir")
	assert.NoError(t, err)
}

func TestOpen(t *testing.T) {
	file, err := xfs.Open("testfile")
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()
}

func TestOpenFile(t *testing.T) {
	file, err := xfs.OpenFile("testfile", os.O_RDWR|os.O_CREATE, 0644)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()
}

func TestResolve(t *testing.T) {
	path, err := xfs.Resolve("testfile", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestRemove(t *testing.T) {
	xfs.EnsureFile("testfile88", 0644)
	err := xfs.Remove("testfile88")
	assert.NoError(t, err)
}

func TestReadFile(t *testing.T) {
	data, err := xfs.ReadFile("testfile")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Equal(t, "test data", string(data))
}

func TestReadTextFile(t *testing.T) {
	data, err := xfs.ReadTextFile("testfile")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Equal(t, "test data", data)
}

func TestReadFileLines(t *testing.T) {
	lines, err := xfs.ReadFileLines("testfile")
	assert.NoError(t, err)
	assert.NotEmpty(t, lines)
	assert.Equal(t, []string{"test data"}, lines)
}

func TestRemoveAll(t *testing.T) {
	xfs.EnsureDir("testdir9999/text", 0755)

	err := xfs.RemoveAll("testdir9999")
	assert.NoError(t, err)
}

func TestRename(t *testing.T) {
	xfs.EnsureFile("testfilex", 0644)
	defer xfs.Remove("testfile_renamed")
	err := xfs.Rename("testfilex", "testfile_renamed")
	assert.NoError(t, err)
	assert.True(t, xfs.Exists("testfile_renamed"))
	assert.False(t, xfs.Exists("testfilex"))
}

func TestStat(t *testing.T) {
	info, err := xfs.Stat("testfile")
	assert.NoError(t, err)
	assert.NotNil(t, info)
}

func TestSymlink(t *testing.T) {
	defer xfs.Remove("testfile_symlink")
	err := xfs.Symlink("testfile", "testfile_symlink")
	assert.NoError(t, err)
}

func TestWalkDir(t *testing.T) {
	err := xfs.WalkDir("testdir", func(path string, d fs.DirEntry, err error) error {
		return nil
	})
	assert.NoError(t, err)
}

func TestWriteFile(t *testing.T) {
	defer xfs.Remove("testfile69")
	err := xfs.WriteFile("testfile69", []byte("test data2"), 0644)
	assert.NoError(t, err)
	data, err := xfs.ReadTextFile("testfile69")
	assert.NoError(t, err)
	assert.Equal(t, "test data2", data)
}

func TestWriteFileLines(t *testing.T) {
	defer xfs.Remove("testfile79")
	err := xfs.WriteFileLines("testfile79", []string{"line1", "line2"}, 0644)
	assert.NoError(t, err)
	data, err := xfs.ReadFileLines("testfile79")
	assert.NoError(t, err)
	assert.Equal(t, []string{"line1", "line2"}, data)
}

func TestWriteFileLinesSep(t *testing.T) {
	defer xfs.Remove("testfile89")
	err := xfs.WriteFileLinesSep("testfile89", []string{"line1", "line2"}, "\n", 0644)
	assert.NoError(t, err)

	data, err := xfs.ReadFileLines("testfile89")
	assert.NoError(t, err)
	assert.Equal(t, []string{"line1", "line2"}, data)
}

func TestWriteTextFile(t *testing.T) {
	defer xfs.Remove("testfile10")
	err := xfs.WriteTextFile("testfile10", "test data10", 0644)
	assert.NoError(t, err)

	data, err := xfs.ReadTextFile("testfile10")
	assert.NoError(t, err)
	assert.Equal(t, "test data10", data)
}
