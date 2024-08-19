// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package flatbuf

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

// / Timestamp is a 64-bit signed integer representing an elapsed time since a
// / fixed epoch, stored in either of four units: seconds, milliseconds,
// / microseconds or nanoseconds, and is optionally annotated with a timezone.
// /
// / Timestamp values do not include any leap seconds (in other words, all
// / days are considered 86400 seconds long).
// /
// / Timestamps with a non-empty timezone
// / ------------------------------------
// /
// / If a Timestamp column has a non-empty timezone value, its epoch is
// / 1970-01-01 00:00:00 (January 1st 1970, midnight) in the *UTC* timezone
// / (the Unix epoch), regardless of the Timestamp's own timezone.
// /
// / Therefore, timestamp values with a non-empty timezone correspond to
// / physical points in time together with some additional information about
// / how the data was obtained and/or how to display it (the timezone).
// /
// /   For example, the timestamp value 0 with the timezone string "Europe/Paris"
// /   corresponds to "January 1st 1970, 00h00" in the UTC timezone, but the
// /   application may prefer to display it as "January 1st 1970, 01h00" in
// /   the Europe/Paris timezone (which is the same physical point in time).
// /
// / One consequence is that timestamp values with a non-empty timezone
// / can be compared and ordered directly, since they all share the same
// / well-known point of reference (the Unix epoch).
// /
// / Timestamps with an unset / empty timezone
// / -----------------------------------------
// /
// / If a Timestamp column has no timezone value, its epoch is
// / 1970-01-01 00:00:00 (January 1st 1970, midnight) in an *unknown* timezone.
// /
// / Therefore, timestamp values without a timezone cannot be meaningfully
// / interpreted as physical points in time, but only as calendar / clock
// / indications ("wall clock time") in an unspecified timezone.
// /
// /   For example, the timestamp value 0 with an empty timezone string
// /   corresponds to "January 1st 1970, 00h00" in an unknown timezone: there
// /   is not enough information to interpret it as a well-defined physical
// /   point in time.
// /
// / One consequence is that timestamp values without a timezone cannot
// / be reliably compared or ordered, since they may have different points of
// / reference.  In particular, it is *not* possible to interpret an unset
// / or empty timezone as the same as "UTC".
// /
// / Conversion between timezones
// / ----------------------------
// /
// / If a Timestamp column has a non-empty timezone, changing the timezone
// / to a different non-empty value is a metadata-only operation:
// / the timestamp values need not change as their point of reference remains
// / the same (the Unix epoch).
// /
// / However, if a Timestamp column has no timezone value, changing it to a
// / non-empty value requires to think about the desired semantics.
// / One possibility is to assume that the original timestamp values are
// / relative to the epoch of the timezone being set; timestamp values should
// / then adjusted to the Unix epoch (for example, changing the timezone from
// / empty to "Europe/Paris" would require converting the timestamp values
// / from "Europe/Paris" to "UTC", which seems counter-intuitive but is
// / nevertheless correct).
// /
// / Guidelines for encoding data from external libraries
// / ----------------------------------------------------
// /
// / Date & time libraries often have multiple different data types for temporal
// / data. In order to ease interoperability between different implementations the
// / Arrow project has some recommendations for encoding these types into a Timestamp
// / column.
// /
// / An "instant" represents a physical point in time that has no relevant timezone
// / (for example, astronomical data). To encode an instant, use a Timestamp with
// / the timezone string set to "UTC", and make sure the Timestamp values
// / are relative to the UTC epoch (January 1st 1970, midnight).
// /
// / A "zoned date-time" represents a physical point in time annotated with an
// / informative timezone (for example, the timezone in which the data was
// / recorded).  To encode a zoned date-time, use a Timestamp with the timezone
// / string set to the name of the timezone, and make sure the Timestamp values
// / are relative to the UTC epoch (January 1st 1970, midnight).
// /
// /  (There is some ambiguity between an instant and a zoned date-time with the
// /   UTC timezone.  Both of these are stored the same in Arrow.  Typically,
// /   this distinction does not matter.  If it does, then an application should
// /   use custom metadata or an extension type to distinguish between the two cases.)
// /
// / An "offset date-time" represents a physical point in time combined with an
// / explicit offset from UTC.  To encode an offset date-time, use a Timestamp
// / with the timezone string set to the numeric timezone offset string
// / (e.g. "+03:00"), and make sure the Timestamp values are relative to
// / the UTC epoch (January 1st 1970, midnight).
// /
// / A "naive date-time" (also called "local date-time" in some libraries)
// / represents a wall clock time combined with a calendar date, but with
// / no indication of how to map this information to a physical point in time.
// / Naive date-times must be handled with care because of this missing
// / information, and also because daylight saving time (DST) may make
// / some values ambiguous or nonexistent. A naive date-time may be
// / stored as a struct with Date and Time fields. However, it may also be
// / encoded into a Timestamp column with an empty timezone. The timestamp
// / values should be computed "as if" the timezone of the date-time values
// / was UTC; for example, the naive date-time "January 1st 1970, 00h00" would
// / be encoded as timestamp value 0.
type Timestamp struct {
	_tab flatbuffers.Table
}

func GetRootAsTimestamp(buf []byte, offset flatbuffers.UOffsetT) *Timestamp {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Timestamp{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Timestamp) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Timestamp) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Timestamp) Unit() TimeUnit {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return TimeUnit(rcv._tab.GetInt16(o + rcv._tab.Pos))
	}
	return 0
}

func (rcv *Timestamp) MutateUnit(n TimeUnit) bool {
	return rcv._tab.MutateInt16Slot(4, int16(n))
}

// / The timezone is an optional string indicating the name of a timezone,
// / one of:
// /
// / * As used in the Olson timezone database (the "tz database" or
// /   "tzdata"), such as "America/New_York".
// / * An absolute timezone offset of the form "+XX:XX" or "-XX:XX",
// /   such as "+07:30".
// /
// / Whether a timezone string is present indicates different semantics about
// / the data (see above).
func (rcv *Timestamp) Timezone() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

// / The timezone is an optional string indicating the name of a timezone,
// / one of:
// /
// / * As used in the Olson timezone database (the "tz database" or
// /   "tzdata"), such as "America/New_York".
// / * An absolute timezone offset of the form "+XX:XX" or "-XX:XX",
// /   such as "+07:30".
// /
// / Whether a timezone string is present indicates different semantics about
// / the data (see above).
func TimestampStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func TimestampAddUnit(builder *flatbuffers.Builder, unit TimeUnit) {
	builder.PrependInt16Slot(0, int16(unit), 0)
}
func TimestampAddTimezone(builder *flatbuffers.Builder, timezone flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(timezone), 0)
}
func TimestampEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}