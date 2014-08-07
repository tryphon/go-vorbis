// Copyright 2012 The Vorbis Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package vorbis contains the Vorbis reference encoder and decoder.
// This is the low-level interface. If you only want to extract audio from
// an Ogg Vorbis file, you probably want to use the vorbisfile package.
package vorbis

/*
#include <stdlib.h>
#include <string.h>
#include <vorbis/codec.h>

// This helper is also used to solve the problem with double array return
// values from C to Go.
void vorbis_set_buffer_item(vorbis_dsp_state *v, int channel, int offset, float value){
	v->pcmret[channel][offset] = value;
}

// This helper is again used to solve the problem with double array return values
// from C to Go.
float vorbis_get_buffer_item(float **pcm, int channel, int offset) {
	return pcm[channel][offset];
}

// This helper is also used to solve the problem with double array return values
// from C to Go.
float *vorbis_get_buffer_indx(float **buf, int channel) {
	return buf[channel];
}

// This helper extract a C string from a array of strings
char *getstring(char **arr_str, int i) {
	return arr_str[i];
}

ogg_packet *ogg_packet_create() {
	return malloc(sizeof(ogg_packet));
}

#cgo LDFLAGS: -lvorbis
*/
import "C"

import (
	ogg "github.com/tryphon/go-ogg"
	"reflect"
	"unsafe"
)

// Info struct contains basic information about the audio in a vorbis bitstream.
type Info C.vorbis_info

func (p *Info) Channels() (channels int16) {
	channels = int16(p.channels)
	return
}

func (p *Info) Rate() (rate int32) {
	rate = int32(p.rate)
	return
}

// DspState struct is the state for one instance of a Vorbis encoder or decoder.
type DspState C.vorbis_dsp_state

func (dspState *DspState) Sequence() int64 {
	return int64(dspState.sequence)
}

func (dspState *DspState) Info() *Info {
	return (*Info)(dspState.vi)
}

// Block struct holds the data for a single block of audio.
type Block C.vorbis_block

// Comment struct defines an Ogg Vorbis comment.
type Comment C.vorbis_comment

func (p *Comment) UserComments() (UserComments []string) {
	if p.user_comments == nil {
		return nil
	}
	UserComments = make([]string, int(p.comments))
	for i := 0; i < int(p.comments); i++ {
		UserComments[i] = C.GoString(C.getstring(p.user_comments, C.int(i)))
	}
	return
}

func (p *Comment) Vendor() string {
	return C.GoString(p.vendor)
}

// Vorbis PRIMITIVES: general

func (p *Info) Init() {
	C.vorbis_info_init((*C.vorbis_info)(p))
}

func (p *Info) Clear() {
	C.vorbis_info_clear((*C.vorbis_info)(p))
}

func (p *Info) BlockSize(zo int16) int16 {
	return int16(C.vorbis_info_blocksize((*C.vorbis_info)(p), C.int(zo)))
}

func (p *Comment) Init() {
	C.vorbis_comment_init((*C.vorbis_comment)(p))
}

func (p *Comment) Add(comment string) {
	cComment := C.CString(comment)
	C.vorbis_comment_add((*C.vorbis_comment)(p), cComment)
	C.free(unsafe.Pointer(cComment))
}

func (p *Comment) AddTag(tag string, contents string) {
	cTag := C.CString(tag)
	cContents := C.CString(contents)
	C.vorbis_comment_add_tag((*C.vorbis_comment)(p), cTag, cContents)
	C.free(unsafe.Pointer(cTag))
	C.free(unsafe.Pointer(cContents))
}

func (p *Comment) Query(tag string, count int) string {
	cTag := C.CString(tag)
	ret := C.vorbis_comment_query((*C.vorbis_comment)(p), cTag, C.int(count))
	C.free(unsafe.Pointer(cTag))
	return C.GoString(ret)
}

func (p *Comment) QueryCount(tag string) int {
	cTag := C.CString(tag)
	ret := C.vorbis_comment_query_count((*C.vorbis_comment)(p), cTag)
	C.free(unsafe.Pointer(cTag))
	return int(ret)
}

func (p *Comment) Clear() {
	C.vorbis_comment_clear((*C.vorbis_comment)(p))
}

func (p *Block) Init(v *DspState) int {
	return int(C.vorbis_block_init((*C.vorbis_dsp_state)(v), (*C.vorbis_block)(p)))
}

func (p *Block) Clear() int {
	return int(C.vorbis_block_clear((*C.vorbis_block)(p)))
}

func (p *DspState) Clear() {
	C.vorbis_dsp_clear((*C.vorbis_dsp_state)(p))
}

func GranuleTime(v *DspState, granulepos int64) float64 {
	return float64(C.vorbis_granule_time((*C.vorbis_dsp_state)(v), C.ogg_int64_t(granulepos)))
}

func VersionString() string {
	return C.GoString(C.vorbis_version_string())
}

// Vorbis PRIMITIVES: analysis/DSP layer

func AnalysisInit(v *DspState, vi *Info) int {
	return int(C.vorbis_analysis_init((*C.vorbis_dsp_state)(v),
		(*C.vorbis_info)(vi)))
}

func (p *Comment) HeaderOut(op *ogg.Packet) int {
	cp := fromPacket(op)
	defer freePacket(cp)
	ret := C.vorbis_commentheader_out((*C.vorbis_comment)(p), cp)
	toPacket(op, cp)
	return int(ret)
}

func AnalysisHeaderOut(v *DspState, vc *Comment, op, op_comm, op_code *ogg.Packet) int {
	cp := fromPacket(op)
	defer freePacket(cp)
	cp_comm := fromPacket(op_comm)
	defer freePacket(cp_comm)
	cp_code := fromPacket(op_code)
	defer freePacket(cp_code)

	ret := int(C.vorbis_analysis_headerout((*C.vorbis_dsp_state)(v),
		(*C.vorbis_comment)(vc), cp, cp_comm, cp_code))

	toPacket(op, cp)
	toPacket(op_comm, cp_comm)
	toPacket(op_code, cp_code)

	return ret
}

func AnalysisSetBufferItem(v *DspState, channel int, offset int, value float32) {
	C.vorbis_set_buffer_item((*C.vorbis_dsp_state)(v), C.int(channel), C.int(offset), C.float(value))
}

func AnalysisBuffer(v *DspState, vals int) [][]float32 {
	floatpp := C.vorbis_analysis_buffer((*C.vorbis_dsp_state)(v), C.int(vals))
	channelCount := int(v.Info().Channels())
	ret := make([][]float32, channelCount)
	for i := 0; i < channelCount; i++ {
		indx := C.vorbis_get_buffer_indx(floatpp, C.int(i))
		h := &reflect.SliceHeader{uintptr(unsafe.Pointer(indx)), vals, vals}
		ret[i] = *(*[]float32)(unsafe.Pointer(h))
	}
	return ret
}

func AnalysisWrote(v *DspState, vals int) int {
	return int(C.vorbis_analysis_wrote((*C.vorbis_dsp_state)(v), C.int(vals)))
}

func AnalysisBlockOut(v *DspState, vb *Block) int {
	return int(C.vorbis_analysis_blockout((*C.vorbis_dsp_state)(v), (*C.vorbis_block)(vb)))
}

func Analysis(vb *Block, op *ogg.Packet) int {
	var cpacket *C.ogg_packet

	if op != nil {
		cpacket = fromPacket(op)
		defer freePacket(cpacket)
	}

	ret := int(C.vorbis_analysis((*C.vorbis_block)(vb), cpacket))

	if op != nil {
		toPacket(op, cpacket)
	}
	return ret
}

func BitrateAddBlock(vb *Block) int {
	return int(C.vorbis_bitrate_addblock((*C.vorbis_block)(vb)))
}

func BitrateFlushPacket(vd *DspState, op *ogg.Packet) int {
	cp := fromPacket(op)
	defer freePacket(cp)
	ret := int(C.vorbis_bitrate_flushpacket((*C.vorbis_dsp_state)(vd), cp))
	toPacket(op, cp)
	return ret
}

// Vorbis PRIMITIVES: synthesis layer

func SynthesisIdHeader(op *ogg.Packet) int {
	cp := fromPacket(op)
	defer freePacket(cp)
	ret := int(C.vorbis_synthesis_idheader(cp))
	toPacket(op, cp)
	return ret
}

func SynthesisHeaderIn(vi *Info, vc *Comment, op *ogg.Packet) int {
	cp := fromPacket(op)
	defer freePacket(cp)
	ret := int(C.vorbis_synthesis_headerin((*C.vorbis_info)(vi),
		(*C.vorbis_comment)(vc), cp))
	toPacket(op, cp)
	return ret
}

func SynthesisInit(v *DspState, vi *Info) int {
	return int(C.vorbis_synthesis_init((*C.vorbis_dsp_state)(v), (*C.vorbis_info)(vi)))
}

func SynthesisRestart(v *DspState) int {
	return int(C.vorbis_synthesis_restart((*C.vorbis_dsp_state)(v)))
}

func Synthesis(vb *Block, op *ogg.Packet) int {
	cp := fromPacket(op)
	defer freePacket(cp)
	ret := int(C.vorbis_synthesis((*C.vorbis_block)(vb), cp))
	toPacket(op, cp)
	return ret
}

func SynthesisTrackOnly(vb *Block, op *ogg.Packet) int {
	cp := fromPacket(op)
	defer freePacket(cp)
	ret := int(C.vorbis_synthesis_trackonly((*C.vorbis_block)(vb), cp))
	toPacket(op, cp)
	return ret
}

func SynthesisBlockin(v *DspState, vb *Block) int {
	return int(C.vorbis_synthesis_blockin((*C.vorbis_dsp_state)(v), (*C.vorbis_block)(vb)))
}

func SynthesisPcmout(v *DspState, pcm ***float32) int {
	return int(C.vorbis_synthesis_pcmout((*C.vorbis_dsp_state)(v), (***C.float)(unsafe.Pointer(pcm))))
}

// This helper function is to solve the problems with C-arrays and Go.
// To avoid these problems, the C-array is just called from within C, for each
// value of the array.
func PcmArrayHelper(pcm **float32, channel, offset int) float32 {
	return float32(C.vorbis_get_buffer_item((**C.float)(unsafe.Pointer(pcm)), C.int(channel), C.int(offset)))
}

func SynthesisLapOut(v *DspState, pcm ***float32) int {
	return int(C.vorbis_synthesis_lapout((*C.vorbis_dsp_state)(v), (***C.float)(unsafe.Pointer(pcm))))
}

func SynthesisRead(v *DspState, samples int) int {
	return int(C.vorbis_synthesis_read((*C.vorbis_dsp_state)(v), C.int(samples)))
}

func PacketBlocksize(vi *Info, op *ogg.Packet) int32 {
	cp := fromPacket(op)
	defer freePacket(cp)
	ret := int32(C.vorbis_packet_blocksize((*C.vorbis_info)(vi), cp))
	toPacket(op, cp)
	return ret
}

func SynthesisHalfrate(v *Info, flag bool) int {
	var cflag C.int
	if flag == true {
		cflag = 1
	}
	return int(C.vorbis_synthesis_halfrate((*C.vorbis_info)(v), cflag))
}

func SynthesisHalfrateP(v *Info) (result bool) {
	ret := int(C.vorbis_synthesis_halfrate_p((*C.vorbis_info)(v)))
	if ret == 1 {
		result = true
	}
	return
}

// Vorbis ERRORS and return codes
const (
	FALSE      = -1
	EOF        = -2
	HOLE       = -3
	EREAD      = -128
	EFAULT     = -129
	EIMPL      = -130
	EINVAL     = -131
	ENOTVORBIS = -132
	EBADHEADER = -133
	EVERSION   = -134
	ENOTAUDIO  = -135
	EBADPACKET = -136
	EBADLINK   = -137
	ENOSEEK    = -138
)

func fromPacket(op *ogg.Packet) *C.ogg_packet {
	if op == nil {
		return nil
	}

	cp := C.ogg_packet_create()

	if op.Packet != nil {
		cp.bytes = C.long(len(op.Packet))
		cp.packet = (*C.uchar)(C.malloc(C.size_t(cp.bytes)))
		C.memcpy(unsafe.Pointer(cp.packet), unsafe.Pointer(&op.Packet[0]), C.size_t(len(op.Packet)))
	}
	if op.BOS {
		cp.b_o_s = 1
	}
	if op.EOS {
		cp.e_o_s = 1
	}
	cp.granulepos = C.ogg_int64_t(op.GranulePos)
	cp.packetno = C.ogg_int64_t(op.PacketNo)
	return cp
}

func toPacket(op *ogg.Packet, cp *C.ogg_packet) {
	if op == nil || cp == nil {
		return
	}

	if cp.packet == nil || cp.bytes == 0 {
		op.Packet = nil
	} else {
		op.Packet = C.GoBytes(unsafe.Pointer(cp.packet), C.int(cp.bytes))
	}

	op.BOS = cp.b_o_s == 1
	op.EOS = cp.e_o_s == 1
	op.GranulePos = int64(cp.granulepos)
	op.PacketNo = int64(cp.packetno)
}

func freePacket(cp *C.ogg_packet) {
	if cp.packet != nil {
		// FIXME
		// C.free(unsafe.Pointer(cp.packet))
	}
	C.free(unsafe.Pointer(cp))
}
