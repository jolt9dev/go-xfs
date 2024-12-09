// Extra fs functions and a drop-in module for many of the file system functions
// in the os module. If you come from other languages, its more intuitive to have
// fs module with the common functions.
//
// The extra functions are Copy, CopyFile, CopyDir, EnsureDir, EnsureFile, ReadTextFile,
// ReadFileLines, WriteTextFile, WriteTextLines
package xfs

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FileMode = os.FileMode

type File = os.File

type FileInfo = os.FileInfo

type DirEntry = fs.DirEntry

// Chown changes the numeric uid and gid of the named file. If the file is a symbolic link,
// it changes the uid and gid of the link's target. A uid or gid of -1 means to not change
// that value. If there is an error, it will be of type [*PathError].
//
// On Windows or Plan 9, Chown always returns the syscall.EWINDOWS or EPLAN9 error, wrapped in *PathError.
//
// Parameters:
//   - filename: the name of the file
//   - uid: the new numeric posix user id
//   - gid: the new numeric posix group id
func Chown(filename string, uid, gid int) error {
	return os.Chown(filename, uid, gid)
}

// Chmod changes the mode of the named file to mode.
// If the file is a symbolic link, it changes the mode of the link's target.
// If there is an error, it will be of type *PathError.
//
// A different subset of the mode bits are used, depending on the
// operating system.
//
// On Unix, the mode's permission bits, ModeSetuid, ModeSetgid, and
// ModeSticky are used.
//
// On Windows, only the 0o200 bit (owner writable) of mode is used; it
// controls whether the file's read-only attribute is set or cleared.
// The other bits are currently unused. For compatibility with Go 1.12
// and earlier, use a non-zero mode. Use mode 0o400 for a read-only
// file and 0o600 for a readable+writable file.
//
// On Plan 9, the mode's permission bits, ModeAppend, ModeExclusive,
// and ModeTemporary are used.
//
// Parameters:
//   - filename: the name of the file
//   - perm: the new file mode e.g. 0644
func Chmod(filename string, perm FileMode) error {
	return os.Chmod(filename, perm)
}

// Copy copies the file from src to dst. The files are only overwritten if the overwrite
// parameter is true. If the file is a symbolic link, it copies the link's target.
//
// Parameters:
//   - src: the source file
//   - dst: the destination file
//   - overwrite: whether to overwrite the destination file if it exists
func Copy(src string, dst string, overwrite bool) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return CopyDir(src, dst, overwrite)
	}

	return CopyFile(src, dst, overwrite)
}

// Copy copies the file from src to dst. The files are only overwritten if the overwrite
// parameter is true. If the file is a symbolic link, it copies the link's target.
//
// Parameters:
//   - src: the source file
//   - dst: the destination file
//   - overwrite: whether to overwrite the destination file if it exists
func CopyDir(src string, dst string, overwrite bool) error {
	return filepath.Walk(src, func(path string, info FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return EnsureDir(dstPath, info.Mode())
		}

		return copyFile(path, dstPath, info, overwrite)
	})
}

// CopyFile copies the file from src to dst. The files are only overwritten if the overwrite
// parameter is true. If the file is a symbolic link, it copies the link's target.
//
// Parameters:
//   - src: the source file
//   - dst: the destination file
//   - overwrite: whether to overwrite the destination file if it exists
func CopyFile(src string, dst string, overwrite bool) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	return copyFile(src, dst, info, overwrite)
}

// Create creates or truncates the named file. If the file already exists, it is truncated.
// If the file does not exist, it is created with mode 0666 (before umask). If successful,
// methods on the returned File can be used for I/O; the associated file descriptor has
// mode O_RDWR.
//
// If there is an error, it will be of type *PathError.
//
// Parameters:
//   - filename: the name of the file
func Create(filename string) (*File, error) {
	return os.Create(filename)
}

// CreateTemp creates a new temporary file in the directory dir with a name beginning with prefix,
// opens the file for reading and writing, and returns the resulting *os.File. If dir is the empty
// string, CreateTemp uses the default directory for temporary files (see os.TempDir). Multiple
// programs calling CreateTemp simultaneously will not choose the same file. The caller can use
// f.Name() to find the pathname of the file. It is the caller's responsibility to remove the file
// when no longer needed.
//
// If there is an error, it will be of type *PathError.
//
// Parameters:
//   - dir: the directory in which to create the file
//   - pattern: the file name pattern
func CreateTemp(dir, pattern string) (*File, error) {
	return os.CreateTemp(dir, pattern)
}

// Getwd returns a rooted path name corresponding to the current directory. If
// the current directory can be reached via multiple paths (due to symbolic links),
// Cwd may return any one of them.
func Cwd() (string, error) {
	return os.Getwd()
}

// Chdir changes the current working directory to the named directory. If there
// is an error, it will be of type *PathError.
//
// Parameters:
//   - dir: the directory to change to
func Chdir(dir string) error {
	return os.Chdir(dir)
}

// Exists reports whether the named file or directory exists.
//
// Parameters:
//   - filename: the name of the file or directory
func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || !os.IsNotExist(err)
}

// EnsureDir creates the named directory with the specified permissions if it does not exist.
//
// Parameters:
//   - dir: the name of the directory
//   - perm: the directory permissions
func EnsureDir(dir string, perm FileMode) error {
	if Exists(dir) {
		return nil
	}

	return os.MkdirAll(dir, perm)
}

// EnsureDirDefault creates the named directory with the default permissions if it does not exist.
//
// Parameters:
//   - dir: the name of the directory
func EnsureDirDefault(dir string) error {
	return EnsureDir(dir, 0755)
}

// EnsureFile creates the named file with the specified permissions if it does not exist.
//
// Parameters:
//   - filename: the name of the file
//   - perm: the file permissions
func EnsureFile(filename string, perm FileMode) error {
	if Exists(filename) {
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	file.Close()
	return os.Chmod(filename, perm)
}

// EnsureFileDefault creates the named file with the default permissions if it does not exist.
//
// Parameters:
//   - filename: the name of the file
func EnsureFileDefault(filename string) error {
	return EnsureFile(filename, 0644)
}

// IsFile reports whether the named file is a file.
//
// Parameters:
//   - filename: the name of the file
func IsFile(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// IsDir reports whether the named file is a directory.
//
// Parameters:
//   - filename: the name of the file
func IsDir(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// IsSymlink reports whether the named file is a symbolic link.
//
// Parameters:
//   - filename: the name of the file
func IsSymlink(filename string) bool {
	info, err := os.Lstat(filename)
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeSymlink != 0
}

// Link creates newname as a hard link to the oldname file. If there is an error, it will be of type *PathError.
//
// Parameters:
//   - oldname: the name of the existing file
//   - newname: the name of the new file
func Link(oldname, newname string) error {
	return os.Link(oldname, newname)
}

// Lstat returns a [FileInfo] describing the named file.
// If the file is a symbolic link, the returned FileInfo
// describes the symbolic link. Lstat makes no attempt to follow the link.
// If there is an error, it will be of type [*PathError].
//
// On Windows, if the file is a reparse point that is a surrogate for another
// named entity (such as a symbolic link or mounted folder), the returned
// FileInfo describes the reparse point, and makes no attempt to resolve it.
//
// Parameters:
//   - filename: the name of the file
func Lstat(filename string) (FileInfo, error) {
	return os.Lstat(filename)
}

// Mkdir creates a new directory with the specified name and permission
// bits (before umask).
//
// If there is an error, it will be of type *PathError.
//
// Parameters:
//   - dir: the name of the directory
//   - perm: the directory permissions
func Mkdir(dir string, perm FileMode) error {
	return os.Mkdir(dir, perm)
}

// MkdirDefault creates a new directory with the specified name and default permissions.
//
// If there is an error, it will be of type *PathError.
//
// Parameters:
//   - dir: the name of the directory
func MkdirDefault(dir string) error {
	return Mkdir(dir, 0755)
}

// MkdirAll creates a directory named path, along with any necessary parents,
// and returns nil, or else returns an error. The permission bits perm (before umask)
// are used for all directories that MkdirAll creates. If path is already a
// directory, MkdirAll does nothing and returns nil.
//
// Parameters:
//   - dir: the name of the directory
//   - perm: the directory permissions
func MkdirAll(dir string, perm FileMode) error {
	return os.MkdirAll(dir, perm)
}

// MkdirAll creates a directory named path, along with any necessary parents,
// and returns nil, or else returns an error. The permission bits perm (before umask)
// are used for all directories that MkdirAll creates. If path is already a
// directory, MkdirAll does nothing and returns nil.
//
// The default permissions are used for the directory (0755).
//
// Parameters:
//   - dir: the name of the directory
//   - perm: the directory permissions
func MkdirAllDefault(dir string) error {
	return MkdirAll(dir, 0755)
}

// Open opens the named file for reading. If successful, methods on the returned file
// can be used for reading; the associated file descriptor has mode O_RDONLY. If there
// is an error, it will be of type *PathError.
//
// Parameters:
//   - filename: the name of the file
func Open(filename string) (*File, error) {
	return os.Open(filename)
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.). If the
// file does not exist, and the O_CREATE flag is passed, it is created with
// mode perm (before umask). If successful, methods on the returned File can
// be used for I/O. If there is an error, it will be of type *PathError.
//
// Parameters:
//   - filename: the name of the file
//   - flag: the file open flag
//   - perm: the file permissions
func OpenFile(filename string, flag int, perm FileMode) (*File, error) {
	return os.OpenFile(filename, flag, perm)
}

// Resolves the relative path to an absolute path. If the relative path is already an absolute path,
// it is returned as is. If the base path is not provided, the current working directory is used.
// If the relative path starts with '~/', the home directory is used as the base path. If the relative
// path starts with './' or '.\', the current working directory is used as the base path. Otherwise,
// the base path is used as the base path.
//
// Parameters:
//   - relative: the relative path
//   - base: the base path
func Resolve(relative string, base string) (string, error) {
	if filepath.IsAbs(relative) {
		return relative, nil
	}

	if base == "" {
		base, _ = Cwd()
	}

	if relative[0] == '~' && (relative[1] == '/' || relative[1] == '\\') {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Abs(filepath.Join(home, relative[2:]))
	}

	if relative[0] == '.' && (relative[1] == '/' || relative[1] == '\\') {
		return filepath.Abs(filepath.Join(base, relative[2:]))
	}

	return filepath.Abs(filepath.Join(base, relative))
}

// Remove removes the named file or (empty) directory. If there is an error, it will be of type *PathError.
//
// Parameters:
//   - filename: the name of the file or directory
func Remove(filename string) error {
	return os.Remove(filename)
}

// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF.
// Because ReadFile reads the whole file, it does not treat an EOF from Read
// as an error to be reported.
//
// Parameters:
//   - filename: the name of the file
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// ReadTextFile reads the named file and returns the contents as a string.
// A successful call returns err == nil, not err == EOF.
// Because ReadTextFile reads the whole file, it does not treat an EOF from Read
// as an error to be reported.
//
// Parameters:
//   - filename: the name of the file
func ReadTextFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// ReadFileLines reads the named file and returns the contents as a slice of lines.
// A successful call returns err == nil, not err == EOF.
// Because ReadFileLines reads the whole file, it does not treat an EOF from Read
// as an error to be reported.
//
// Parameters:
//   - filename: the name of the file
func ReadFileLines(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error
// it encounters. If the path does not exist, RemoveAll
// returns nil (no error).
// If there is an error, it will be of type [*PathError].
//
// Parameters:
//   - path: the name of the file or directory
func RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Rename renames (moves) oldpath to newpath.
// If newpath already exists and is not a directory, Rename replaces it.
// OS-specific restrictions may apply when oldpath and newpath are in different directories.
// Even within the same directory, on non-Unix platforms Rename is not an atomic operation.
// If there is an error, it will be of type *LinkError.
//
// Parameters:
//   - oldpath: the old name of the file or directory
func Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Stat returns a [FileInfo] describing the named file.
// If there is an error, it will be of type [*PathError].
//
// Parameters:
//   - filename: the name of the file
func Stat(filename string) (FileInfo, error) {
	return os.Stat(filename)
}

// Symlink creates newname as a symbolic link to oldname.
// On Windows, a symlink to a non-existent oldname creates a file symlink;
// if oldname is later created as a directory the symlink will not work.
// If there is an error, it will be of type *LinkError.
//
// Parameters:
//   - oldname: the name of the existing file
func Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// WalkDir walks the file tree rooted at root, calling fn for each file or
// directory in the tree, including root.
//
// All errors that arise visiting files and directories are filtered by fn:
// see the [fs.WalkDirFunc] documentation for details.
//
// The files are walked in lexical order, which makes the output deterministic
// but requires WalkDir to read an entire directory into memory before proceeding
// to walk that directory.
//
// WalkDir does not follow symbolic links.
//
// WalkDir calls fn with paths that use the separator character appropriate
// for the operating system. This is unlike [io/fs.WalkDir], which always
// uses slash separated paths.
//
// Parameters:
//   - root: the root directory
//   - walkFn: the walk function
func WalkDir(root string, walkFn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, walkFn)
}

// WriteFile writes data to the named file, creating it if necessary.
// If the file does not exist, WriteFile creates it with permissions perm (before umask);
// otherwise WriteFile truncates it before writing, without changing permissions.
// Since WriteFile requires multiple system calls to complete, a failure mid-operation
// can leave the file in a partially written state.
//
// Parameters:
//   - filename: the name of the file
//   - data: the data to write
//   - perm: the file permissions
func WriteFile(filename string, data []byte, perm FileMode) error {
	return os.WriteFile(filename, data, perm)
}

// WriteFileLines writes the lines to the named file, creating it if necessary.
// If the file does not exist, WriteFileLines creates it with permissions perm (before umask);
// otherwise WriteFileLines truncates it before writing, without changing permissions.
// Since WriteFileLines requires multiple system calls to complete, a failure mid-operation
// can leave the file in a partially written state.
//
// The lines are separated by the default end of line character for the platform.
//
// Parameters:
//   - filename: the name of the file
//   - lines: the lines to write
//   - perm: the file permissions
func WriteFileLines(filename string, lines []string, perm FileMode) error {
	return WriteFileLinesSep(filename, lines, EOL, perm)
}

// WriteFileLines writes the lines to the named file, creating it if necessary.
// If the file does not exist, WriteFileLines creates it with permissions perm (before umask);
// otherwise WriteFileLines truncates it before writing, without changing permissions.
// Since WriteFileLines requires multiple system calls to complete, a failure mid-operation
// can leave the file in a partially written state.
//
// Parameters:
//   - filename: the name of the file
//   - lines: the lines to write
//   - sep: the line separator
//   - perm: the file permissions
func WriteFileLinesSep(filename string, lines []string, sep string, perm FileMode) error {
	sb := strings.Builder{}
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteString(sep)
	}

	return WriteTextFile(filename, sb.String(), perm)
}

// WriteTextFile writes the text to the named file, creating it if necessary.
// If the file does not exist, WriteTextFile creates it with permissions perm (before umask);
// otherwise WriteTextFile truncates it before writing, without changing permissions.
// Since WriteTextFile requires multiple system calls to complete, a failure mid-operation
// can leave the file in a partially written state.
//
// Parameters:
//   - filename: the name of the file
//   - data: the text to write
//   - perm: the file permissions
func WriteTextFile(filename string, data string, perm FileMode) error {
	return os.WriteFile(filename, []byte(data), perm)
}

func copyFile(src, dst string, info FileInfo, overwrite bool) error {

	if Exists(dst) && !overwrite {
		return nil
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, info.Mode())
}

// WalkDirFunc is the type of the function called by WalkDir to visit each file or directory.
// The path argument contains the argument to WalkDir as a prefix; that is, if WalkDir is
// called with "dir", which is a directory containing the file "a", the walk function will
// be called with argument "dir/a". The info argument is the os.FileInfo for the named path.
// The error result is nil if the call succeeds.
type WalkDirFunc func(path string, d DirEntry, err error) error
