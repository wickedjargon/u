package main

import (
	"bytes"
	"encoding/binary"

	exif "github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jis "github.com/dsoprea/go-jpeg-image-structure/v2"
)

// Pyongyang, North Korea — 39°1'27.48"N, 125°45'12.6"E.
// Built as raw EXIF rationals so the fractional seconds survive (GpsDegrees.Raw
// truncates seconds to a whole number).
var (
	pyongyangLat = []exifcommon.Rational{
		{Numerator: 39, Denominator: 1},
		{Numerator: 1, Denominator: 1},
		{Numerator: 2748, Denominator: 100},
	}
	pyongyangLon = []exifcommon.Rational{
		{Numerator: 125, Denominator: 1},
		{Numerator: 45, Denominator: 1},
		{Numerator: 1260, Denominator: 100},
	}
)

// trollExif strips all existing EXIF from a JPEG and replaces it with a GPS tag
// pointing at Pyongyang. Anything that isn't a JPEG, or that trips up the EXIF
// libraries, is returned unchanged — the upload must never fail because of the
// joke.
func trollExif(data []byte) (result []byte) {
	result = data

	// JPEG SOI marker. The dsoprea parser only handles JPEG; everything else
	// passes through untouched.
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		return data
	}

	// The dsoprea libraries panic on malformed input; never let that escape.
	defer func() {
		if recover() != nil {
			result = data
		}
	}()

	jmp := jis.NewJpegMediaParser()
	intfc, err := jmp.ParseBytes(data)
	if err != nil {
		return data
	}
	sl := intfc.(*jis.SegmentList)

	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		return data
	}
	ti := exif.NewTagIndex()

	// Fresh root builder => the rewritten EXIF segment contains only what we add
	// below, so any original GPS (and other metadata) is dropped.
	rootIb := exif.NewIfdBuilder(im, ti, exifcommon.IfdStandardIfdIdentity, binary.BigEndian)

	gpsIb, err := exif.GetOrCreateIbFromRootIb(rootIb, "IFD/GPSInfo")
	if err != nil {
		return data
	}

	for _, t := range []struct {
		name string
		val  interface{}
	}{
		{"GPSLatitudeRef", "N"},
		{"GPSLatitude", pyongyangLat},
		{"GPSLongitudeRef", "E"},
		{"GPSLongitude", pyongyangLon},
	} {
		if err := gpsIb.SetStandardWithName(t.name, t.val); err != nil {
			return data
		}
	}

	if err := sl.SetExif(rootIb); err != nil {
		return data
	}

	out := new(bytes.Buffer)
	if err := sl.Write(out); err != nil {
		return data
	}
	return out.Bytes()
}
