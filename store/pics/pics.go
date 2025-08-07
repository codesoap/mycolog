// Package pics deals with the storage of pictures. The pictures are
// related to components and their filenames are composed like this:
//
//	<component-ID>_<position>.<extension>
//
// Picture's position may be changed by changing <position>. They
// may also be deleted from the filesystem.
package pics

// TODO: Think about a cleanup for pictures with missing components.
// TODO: Allow changing the position of pictures.

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const maxImgSize = 10 * 1024 * 1024 // 10 MiB
const maxResolution = 10_000        // Max 10k pixels width or height.

type PictureStore struct {
	Path string // Path is the path to the picture directory.
}

// PictureName is the filename of a picture inside a PictureStore.
// It contains only the filename without any directories.
type PictureName string

// Add adds a new picture for the component with componentID. It will be
// added at the next free position. An error will be returned, if the
// picture has not been added; this can happen if the picture does not
// meet some validation criteria.
func (s PictureStore) Add(componentID int64, rawPic io.Reader) (PictureName, error) {
	limitedReader := io.LimitReader(rawPic, maxImgSize+1)
	picsBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("could not read picture: %s", err)
	} else if len(picsBytes) > maxImgSize {
		return "", fmt.Errorf("picture is more than 10 MiB large")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(picsBytes))
	if err := validate(cfg, err); err != nil {
		return "", fmt.Errorf("invalid picture: %s", err)
	}

	pics := s.Pictures(componentID)
	name := fmt.Sprintf("%d_%d.%s", componentID, len(pics), format)
	path := filepath.Join(s.Path, name)
	return PictureName(name), os.WriteFile(path, picsBytes, 0600)
}

func validate(cfg image.Config, err error) error {
	if err != nil {
		return err
	} else if cfg.Height > maxResolution {
		return fmt.Errorf("picture is higher than %d pixels", maxResolution)
	} else if cfg.Width > maxResolution {
		return fmt.Errorf("picture is wider than %d pixels", maxResolution)
	}
	return nil
}

// Pictures returns picture names of pictures for the component with
// componentID.
func (s PictureStore) Pictures(componentID int64) []PictureName {
	paths, err := filepath.Glob(path.Join(s.Path, fmt.Sprintf("%d_*", componentID)))
	if err != nil {
		return nil
	}
	filenames := make([]PictureName, len(paths))
	for i, p := range paths {
		filenames[i] = PictureName(filepath.Base(p))
	}
	return filenames
}

// Delete deletes the given picture from the filesystem. Pictures for
// the same component with a higher position will have their position
// decremented.
func (s PictureStore) Delete(pictureName PictureName) error {
	componentID := strings.Split(string(pictureName), "_")[0]
	id, err := strconv.ParseInt(componentID, 10, 64)
	if err != nil {
		format := "could not determine component of picture '%s'"
		return fmt.Errorf(format, pictureName)
	}
	err = os.Remove(filepath.Join(s.Path, string(pictureName)))
	if err != nil {
		return err
	}
	pics := s.Pictures(id)
	sort.Slice(pics, func(i, j int) bool {
		return strings.Compare(string(pics[i]), string(pics[j])) == -1
	})
	for i, pic := range pics {
		expPrefix := fmt.Sprintf("%d_%d.", id, i)
		if !strings.HasPrefix(string(pic), expPrefix) {
			ext := filepath.Ext(string(pic))
			old := filepath.Join(s.Path, string(pic))
			new := filepath.Join(s.Path, expPrefix[:len(expPrefix)-1]+ext)
			if err = os.Rename(old, new); err != nil {
				format := "could not update file '%s' to '%s' name: %s"
				return fmt.Errorf(format, old, new, err)
			}

			// Touch file to avoid something like a web browser not updating
			// caches:
			_ = os.Chtimes(new, time.Time{}, time.Now())
		}
	}
	return nil
}

// DeleteForComponent deletes all pictures of the component with
// componentID.
func (s PictureStore) DeleteForComponent(componentID int64) error {
	paths, err := filepath.Glob(path.Join(s.Path, fmt.Sprintf("%d_*", componentID)))
	if err != nil {
		return err
	}
	for _, p := range paths {
		if err = os.Remove(p); err != nil {
			return fmt.Errorf("could not remove '%s': %s", p, err)
		}
	}
	return nil
}
