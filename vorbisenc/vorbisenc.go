// Copyright 2012 The Vorbis Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vorbisenc

/*
#cgo LDFLAGS: -lvorbisenc
#include <vorbis/vorbisenc.h>
*/
import "C"

import (
	"unsafe"

	"github.com/grd/vorbis"
)

func Init(vi *vorbis.Info, channels int32, rate int32,
	max_bitrate int32, nominal_bitrate int32, min_bitrate int32) int {
	return int(C.vorbis_encode_init((*C.vorbis_info)(unsafe.Pointer(vi)),
		C.long(channels), C.long(rate), C.long(max_bitrate),
		C.long(nominal_bitrate), C.long(min_bitrate)))
}

func SetupManaged(vi *vorbis.Info, channels int32, rate int32,
	max_bitrate int32, nominal_bitrate int32, min_bitrate int32) int {
	return int(C.vorbis_encode_setup_managed((*C.vorbis_info)(unsafe.Pointer(vi)),
		C.long(channels), C.long(rate), C.long(max_bitrate),
		C.long(nominal_bitrate), C.long(min_bitrate)))
}

func SetupVbr(vi *vorbis.Info, channels int32, rate int32, quality float32) int {
	return int(C.vorbis_encode_setup_vbr((*C.vorbis_info)(unsafe.Pointer(vi)),
		C.long(channels), C.long(rate), C.float(quality)))
}

func InitVbr(vi *vorbis.Info, channels int32, rate int32, base_quality float32) int {
	return int(C.vorbis_encode_init_vbr((*C.vorbis_info)(unsafe.Pointer(vi)),
		C.long(channels), C.long(rate), C.float(base_quality)))
}

func SetupInit(vi *vorbis.Info) int {
	return int(C.vorbis_encode_setup_init((*C.vorbis_info)(unsafe.Pointer(vi))))
}

func Ctl(vi *vorbis.Info, number int, arg *uintptr) int {
	return int(C.vorbis_encode_ctl((*C.vorbis_info)(unsafe.Pointer(vi)),
		C.int(number), unsafe.Pointer(arg)))
}
