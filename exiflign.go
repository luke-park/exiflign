// This package exposes an interface for "normalizing" JPEG images that have
// their orientation EXIF encoded.  This library was designed for working with
// images uploaded from phone cameras that usually have their orientation
// tagged, which results in a rotated/mirrored images when using the
// Go image/jpeg library.  Supports little-endian and big-endian EXIF encodings,
// as well as all possible tag transformations.
package exiflign

import (
	"errors"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
)

const bufferSize = 32

var bufferExif = []int{0x45, 0x78, 0x69, 0x66, 0x00, 0x00}
var bufferOrienLittle = []int{0x12, 0x01, 0x03, 0x00, 0x01, 0x00, 0x00, 0x00, -1, 0x00}
var bufferOrienBig = []int{0x01, 0x12, 0x00, 0x03, 0x00, 0x00, 0x00, 0x01, 0x00, -1 }

var NoExifError error = errors.New("The given file does not contain any EXIF orientation information.")

// Normalize is the "quick-fix" function of this package.  It requires an
// io.ReadSeeker since it needs to detect the orientation and then decode the
// image.  It will write the orientation-corrected image to w.  If the JPEG
// image in r does not have EXIF data or does not have orientation data, r is
// simply copied to w.  When finished, the internal position in r will be at
// io.SeekStart.
func Normalize(r io.ReadSeeker, w io.Writer) error {
	tag, err := GetOrientationTag(r)
	if err == NoExifError {
		_, err = io.Copy(w, r)
		if err != nil {
			return err
		}
	}

	img1, err := jpeg.Decode(r)
	if err != nil {
		return err
	}

	img2 := TransformForTag(img1, tag)
	err = jpeg.Encode(w, img2, nil)
	if err != nil {
		return err
	}

	return nil
}

// TransformForTag performs the neccessary transformation on img that will
// facilitate removal of the orientation tag.
func TransformForTag(img image.Image, tag uint16) image.Image {
	switch tag {
	default:
		return img
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipH(imaging.Rotate180(img))
	case 5:
		return imaging.FlipH(imaging.Rotate270(img))
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.FlipH(imaging.Rotate90(img))
	case 8:
		return imaging.Rotate90(img)
	}
}

// GetOrientationTag produces a value between 1 and 8, inclusive, for a given
// JPEG image in r.  This value describes the transformations required to
// produce the correct image.  The excellent article by Magnus Hoff covers this
// in more detail:
//
// https://magnushoff.com/jpeg-orientation.html
func GetOrientationTag(r io.ReadSeeker) (uint16, error) {
	endr, err := splitSearch(r, bufferExif, 8)
	if err != nil {
		return 0, NoExifError
	}
	r.Seek(0, io.SeekStart)

	littleEndian := endr[6] == 0x49 && endr[7] == 0x49
	bufferOrien := bufferOrienBig
	if littleEndian {
		bufferOrien = bufferOrienLittle
	}

	res, err := splitSearch(r, bufferOrien, 10)
	if err != nil {
		return 0, NoExifError
	}
	r.Seek(0, io.SeekStart)

	tagr := res[8:]
	if littleEndian {
		tagr[0], tagr[1] = tagr[1], tagr[0]
	}

	var tag uint16
	tag |= uint16(tagr[1])
	tag += uint16(tagr[0]) << 8

	if tag < 1 || tag > 8 {
		tag = 1
	}

	return tag, nil
}
func splitSearch(r io.ReadSeeker, sequence []int, length uint) ([]byte, error) {
	var buffer [bufferSize]byte
	v1 := buffer[:bufferSize/2]
	v2 := buffer[bufferSize/2:]

	_, err := r.Read(buffer[:])
	if err != nil {
		return nil, err
	}

	var res []byte = find(buffer, sequence, length)
	if res != nil {
		return res, nil
	}

	for res = nil; res == nil; res = find(buffer, sequence, length) {
		copy(v1, v2)

		_, err := r.Read(v2)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
func find(buffer [bufferSize]byte, sequence []int, length uint) []byte {
	for i := 0; i <= len(buffer)-int(length); i++ {
		didMatch := true

		for j := 0; j < len(sequence); j++ {
			if buffer[i+j] != byte(sequence[j]) && sequence[j] != -1 {
				didMatch = false
				break
			}
		}

		if didMatch {
			return buffer[i : i+int(length)]
		}
	}

	return nil
}
