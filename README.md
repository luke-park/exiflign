![](icon.png)

<a href="https://godoc.org/github.com/luke-park/exiflign"><img src="https://godoc.org/github.com/luke-park/exiflign?status.svg" alt="GoDoc"></a>

# exiflign
This package exposes an interface for "normalizing" JPEG images that have
their orientation EXIF encoded.  This library was designed for working with
images uploaded from phone cameras that usually have their orientation
tagged, which results in rotated/mirrored images when using the
Go image/jpeg library.  Supports little-endian and big-endian EXIF encodings,
as well as all possible tag transformations.

## Example
This package is very easy to use, but exposes lower-level functions for more
control.  A very basic use example:
```go
fIn, err := os.Open("./input.jpg")
if err != nil { return err }
defer fIn.Close()

fOut, err := os.Create("./output.jpg")
if err != nil { return err }
defer fOut.Close()

err = exiflign.Normalize(fIn, fOut)
if err != nil { return err }
```

More control, as well as in-memory transformations, can also be performed.

## Documentation
The full documentation of this package can be found on [GoDoc](https://godoc.org/github.com/luke-park/exiflign).
