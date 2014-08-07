// Copyright 2012 The Vorbis Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Function: simple example encoder
// Takes a stereo 16bit 44.1kHz WAV file from stdin and encodes it into
// a Vorbis bitstream
package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"

	ogg "github.com/tryphon/go-ogg"
	vorbis "github.com/tryphon/go-vorbis"
	"github.com/tryphon/go-vorbis/vorbisenc"
)

const READ = 1024

var readbuffer = make([]byte, READ*4+44)

func main() {
	var (
		oss ogg.StreamState // take physical pages, weld into a logical stream of packets
		og  ogg.Page        // one Ogg bitstream page. Vorbis packets are inside
		op  ogg.Packet      // one raw packet of data for decode

		vi vorbis.Info     // struct that stores all the static vorbis bitstream settings
		vc vorbis.Comment  // struct that stores all the user comments
		vd vorbis.DspState // central working state for the packet PCM decoder
		vb vorbis.Block    // local working space for packet PCM decode

		ret, i, val int
		eos         bool
	)

	// we cheat on the WAV header; we just bypass the header and never
	// verify that it matches 16bit/stereo/44.1kHz. This is just an
	// example, after all.

	for i < 30 {
		val, _ = os.Stdin.Read(readbuffer[0:2])
		if val < 2 {
			break
		}

		if bytes.Equal(readbuffer[0:2], []byte("da")) {
			val, _ = os.Stdin.Read(readbuffer[0:6])
			break
		}
		i++
	}

	// Encode setup

	vi.Init()

	/* choose an encoding mode.  A few possibilities commented out, one
	 actually used:

	*********************************************************************
	Encoding using a VBR quality mode.  The usable range is -.1
	(lowest quality, smallest file) to 1. (highest quality, largest file).
	Example quality mode .4: 44kHz stereo coupled, roughly 128kbps VBR

	ret = vorbisenc.InitVbr(&vi,2,44100,.4);

	---------------------------------------------------------------------

	Encoding using an average bitrate mode (ABR).
	example: 44kHz stereo coupled, average 128kbps VBR

	ret = vorbisenc.Init(&vi,2,44100,-1,128000,-1);

	---------------------------------------------------------------------

	Encode using a quality mode, but select that quality mode by asking for
	an approximate bitrate.  This is not ABR, it is true VBR, but selected
	using the bitrate interface, and then turning bitrate management off:

	ret = ( vorbisenc.SetupManaged(&vi,2,44100,-1,128000,-1) ||
	        vorbisenc.Ctl(&vi,vorbis.ECTL_RATEMANAGE2_SET,nil) ||
	        vorbisenc.SetupInit(&vi));

	*********************************************************************/

	ret = vorbisenc.InitVbr(&vi, 2, 44100, 0.1)

	// do not continue if setup failed; this can happen if we ask for a
	// mode that libVorbis does not support (eg, too low a bitrate, etc,
	// will return 'EIMPL')

	if ret != 0 {
		os.Exit(1)
	}

	// add a comment
	vc.Init()
	vc.AddTag("ENCODER", "encoder_example.go")

	// set up the analysis state and auxiliary encoding storage
	vorbis.AnalysisInit(&vd, &vi)
	vb.Init(&vd)

	// set up our packet stream encoder (oss)
	// pick a random serial number; that way we can more likely build
	// chained streams just by concatenation
	oss.Init(rand.Int31())

	// Vorbis streams begin with three headers; the initial header (with
	// most of the codec setup parameters) which is mandated by the Ogg
	// bitstream spec.  The second header holds any comment fields.  The
	// third header holds the bitstream codebook.  We merely need to
	// make the headers, then pass them to the vorbis package one at a time;
	// Vorbis handles the additional Ogg bitstream constraints

	var (
		header     ogg.Packet
		headerComm ogg.Packet
		headerCode ogg.Packet
	)

	vorbis.AnalysisHeaderOut(&vd, &vc, &header, &headerComm, &headerCode)
	oss.PacketIn(&header)
	oss.PacketIn(&headerComm)
	oss.PacketIn(&headerCode)

	// This ensures the actual audio data will start on a new page, as per spec

	for {
		result := oss.Flush(&og)
		if result == false {
			break
		}
		os.Stdout.Write(og.Header)
		os.Stdout.Write(og.Body)
	}

	for !eos {
		var i int
		ret, _ = os.Stdin.Read(readbuffer[0 : READ*4]) // stereo hardwired here

		if ret == 0 {
			// end of file.  this can be done implicitly in the mainline,
			// but it's easier to see here in non-clever fashion.
			// Tell the library we're at end of stream so that it can handle
			// the last frame and mark end of stream in the output properly
			vorbis.AnalysisWrote(&vd, 0)

		} else {
			// data to encode

			// expose the buffer to submit data
			buffer := vorbis.AnalysisBuffer(&vd, READ)

			// uninterleave samples
			for i = 0; i < (ret / 4); i++ {
				buffer[0][i] = float32((int16(readbuffer[i*4+1])<<8)|
					(0x00ff&int16(readbuffer[i*4]))) / 32768.
				buffer[1][i] = float32((int16(readbuffer[i*4+3])<<8)|
					(0x00ff&int16(readbuffer[i*4+2]))) / 32768.
			}

			// tell the library how much we actually submitted
			vorbis.AnalysisWrote(&vd, i)
		}

		// Vorbis does some data preanalysis, then divvies up blocks for
		// more involved (potentially parallel) processing.  Get a single
		// block for encoding now
		for vorbis.AnalysisBlockOut(&vd, &vb) == 1 {

			// analysis, assume we want to use bitrate management
			vorbis.Analysis(&vb, nil)
			vorbis.BitrateAddBlock(&vb)

			for vorbis.BitrateFlushPacket(&vd, &op) != 0 {

				// weld the packet into the bitstream
				oss.PacketIn(&op)

				// write out pages (if any)
				for !eos {
					if oss.PageOut(&og) == false {
						break
					}

					os.Stdout.Write(og.Header)
					os.Stdout.Write(og.Body)

					// this could be set above, but for illustrative purposes, I do
					// it here (to show that vorbis does know where the stream ends)
					eos = og.Eos()
				}
			}
		}
	}

	// clean up and exit. vi.Clear() must be called last

	vb.Clear()
	vd.Clear()
	vc.Clear()
	vi.Clear()

	// ogg.Page and ogg.Packet structs always point to storage in
	// lib.vorbis. They're never freed or manipulated directly

	fmt.Fprintf(os.Stderr, "Done.\n")
	os.Exit(0)
}
