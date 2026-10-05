// Package settingsfile keeps a ribbon's settings in one indented JSON file a person can read.
//
// Each application says what its file holds through a Codec; this package holds what every ribbon's
// file does alike. Reading is tolerant: one bad value leaves its default and one bad entry of a list
// never costs the rest. A file that cannot be trusted at all is kept aside under another name, never
// overwritten; a file that is there but cannot be read is never saved over. Keys a later version
// wrote are written back as found. Writing replaces the file whole through atomicfile.
package settingsfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/oernster/ribbonkit/infrastructure/atomicfile"
)

// File names inside the settings folder.
const (
	FileName           = "settings.json"
	keptAsideStem      = "settings.unreadable"
	keptAsideExtension = ".json"
	UnreadableName     = keptAsideStem + keptAsideExtension
)

// KeptAsideLimit bounds how many damaged files are kept aside. Once every name is taken the next is
// not renamed and saving is refused, so even then nothing is overwritten.
const KeptAsideLimit = 100

// Permissions for what a store creates: the folder and the file are the user's alone.
const (
	FolderMode fs.FileMode = 0o700
	FileMode   fs.FileMode = 0o600
)

// KeptAsideNotice is shown when an unreadable file was renamed to name.
func KeptAsideNotice(name string) string {
	return "Settings could not be read; the old file was kept as " + name
}

// KeptAsideName answers the name the nth kept-aside file takes: UnreadableName for the first, then
// the same name numbered from 2, so an earlier copy is never the one replaced.
func KeptAsideName(n int) string {
	if n <= 1 {
		return UnreadableName
	}
	return keptAsideStem + "-" + strconv.Itoa(n) + keptAsideExtension
}

// ErrNotRead is answered by Load and then by every Save when the file was there but could not be
// read: held open by another program, say. The defaults standing in for it are not the user's, so
// nothing is saved over it until the next run reads it.
var ErrNotRead = errors.New("nothing is saved over it")

// ErrNotKeptAside is answered by Save when an unreadable file could not be renamed, so writing
// would destroy the only copy of the user's settings.
var ErrNotKeptAside = errors.New("the unreadable settings file could not be kept aside, so it is not overwritten")

// errNoFreeName is answered when every kept-aside name is taken.
var errNoFreeName = errors.New("every name a damaged file is kept under is taken")

// Codec is what one application's file holds.
type Codec[T any] struct {
	// Keys are the top-level keys this version reads, in the order they are written.
	Keys []string
	// Defaults answers the settings of a first run.
	Defaults func() T
	// Decode reads the settings from a file's object, starting from the defaults. It answers false
	// when the file cannot be trusted at all, which keeps it aside.
	Decode func(Object) (T, bool)
	// Values answers the value of every one of Keys for settings; a missing key is a fault.
	Values func(T) (map[string]any, error)
}

// Store keeps one application's settings file in one folder.
type Store[T any] struct {
	dir     string
	product string
	codec   Codec[T]
	// extras are top-level keys this version does not know, written back as found.
	extras []member
	// blocked is why saving is refused: ErrNotRead or ErrNotKeptAside beneath; nil while it is not.
	blocked error
}

// New answers a store over dir for product, the name the user knows it by, with codec. Nothing is
// read or made until Load or Save is called.
func New[T any](dir, product string, codec Codec[T]) *Store[T] {
	return &Store[T]{dir: dir, product: product, codec: codec}
}

// Path answers the settings file's path.
func (s *Store[T]) Path() string { return filepath.Join(s.dir, FileName) }

// Extras answers the top-level keys the last Load found that this version does not read.
func (s *Store[T]) Extras() []string {
	keys := make([]string, 0, len(s.extras))
	for _, each := range s.extras {
		keys = append(keys, each.key)
	}
	return keys
}

// Load reads the settings with the notice the user should read; empty when none. No file answers
// the defaults with no notice. A file that cannot be trusted is kept aside and the defaults are
// answered with the notice saying where it went. A fault reading a file that is there is answered as
// an error and refuses every later save.
func (s *Store[T]) Load() (T, string, error) {
	raw, err := os.ReadFile(s.Path())
	if errors.Is(err, fs.ErrNotExist) {
		return s.codec.Defaults(), "", nil
	}
	if err != nil {
		s.blocked = fmt.Errorf("reading %s: %w; %w until %s is started again and reads it", s.Path(), err, ErrNotRead, s.product)
		return s.codec.Defaults(), "", s.blocked
	}
	object, ok := parse(raw)
	if !ok {
		return s.keepAside()
	}
	decoded, ok := s.codec.Decode(object)
	if !ok {
		return s.keepAside()
	}
	s.extras = extrasOf(raw, object, s.codec.Keys)
	return decoded, "", nil
}

// keepAside renames an unreadable file to the first kept-aside name not yet taken and answers the
// defaults with the notice saying where it went; an earlier copy is never replaced. If no name is
// free or the rename fails, saving is refused from then on.
func (s *Store[T]) keepAside() (T, string, error) {
	name, err := s.freeKeptAsideName()
	if err == nil {
		err = os.Rename(s.Path(), filepath.Join(s.dir, name))
	}
	if err != nil {
		s.blocked = fmt.Errorf("%w: %w", ErrNotKeptAside, err)
		return s.codec.Defaults(), "", s.blocked
	}
	return s.codec.Defaults(), KeptAsideNotice(name), nil
}

// freeKeptAsideName answers the first kept-aside name nothing in the folder holds.
func (s *Store[T]) freeKeptAsideName() (string, error) {
	for n := 1; n <= KeptAsideLimit; n++ {
		name := KeptAsideName(n)
		_, err := os.Lstat(filepath.Join(s.dir, name))
		if errors.Is(err, fs.ErrNotExist) {
			return name, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", errNoFreeName
}

// Save writes the settings, making the folder where it is not there. It replaces the file whole, so
// a failure part way leaves the previous file as it was.
func (s *Store[T]) Save(current T) error {
	if s.blocked != nil {
		return s.blocked
	}
	values, err := s.codec.Values(current)
	if err != nil {
		return err
	}
	body, err := encode(s.codec.Keys, values, s.extras)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(s.dir, FolderMode); err != nil {
		return fmt.Errorf("making %s: %w", s.dir, err)
	}
	return atomicfile.Write(s.Path(), body, FileMode)
}
