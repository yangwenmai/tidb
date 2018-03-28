// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	. "github.com/pingcap/check"
	"github.com/pingcap/tidb/mysql"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/testleak"
)

var _ = Suite(&testUtilSuite{})

type testUtilSuite struct {
}

func (s *testUtilSuite) TestDumpBinaryTime(c *C) {
	defer testleak.AfterTest(c)()
	t, err := types.ParseTimestamp(nil, "0000-00-00 00:00:00.0000000")
	c.Assert(err, IsNil)
	d, err := dumpBinaryDateTime(nil, t, nil)
	c.Assert(err, IsNil)
	c.Assert(d, DeepEquals, []byte{11, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0})
	t, err = types.ParseDatetime(nil, "0000-00-00 00:00:00.0000000")
	c.Assert(err, IsNil)
	d, err = dumpBinaryDateTime(nil, t, nil)
	c.Assert(err, IsNil)
	c.Assert(d, DeepEquals, []byte{11, 1, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0})

	t, err = types.ParseDate(nil, "0000-00-00")
	c.Assert(err, IsNil)
	d, err = dumpBinaryDateTime(nil, t, nil)
	c.Assert(err, IsNil)
	c.Assert(d, DeepEquals, []byte{4, 1, 0, 1, 1})

	myDuration, err := types.ParseDuration("0000-00-00 00:00:00.0000000", 6)
	c.Assert(err, IsNil)
	d = dumpBinaryTime(myDuration.Duration)
	c.Assert(d, DeepEquals, []byte{0})
}

func (s *testUtilSuite) TestDumpTextValue(c *C) {
	defer testleak.AfterTest(c)()

	columns := []*ColumnInfo{{
		Type:    mysql.TypeLonglong,
		Decimal: mysql.NotFixedDec,
	}}
	bs, err := dumpTextRow(nil, columns, types.DatumRow{types.NewIntDatum(10)})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "10")

	bs, err = dumpTextRow(nil, columns, types.DatumRow{types.NewUintDatum(11)})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "11")

	columns[0].Type = mysql.TypeFloat
	columns[0].Decimal = 1
	f32 := types.NewFloat32Datum(1.2)
	bs, err = dumpTextRow(nil, columns, types.DatumRow{f32})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "1.2")

	columns[0].Decimal = 2
	bs, err = dumpTextRow(nil, columns, types.DatumRow{f32})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "1.20")

	f64 := types.NewFloat64Datum(2.2)
	columns[0].Type = mysql.TypeDouble
	columns[0].Decimal = 1
	bs, err = dumpTextRow(nil, columns, types.DatumRow{f64})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "2.2")

	columns[0].Decimal = 2
	bs, err = dumpTextRow(nil, columns, types.DatumRow{f64})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "2.20")

	columns[0].Type = mysql.TypeBlob
	bs, err = dumpTextRow(nil, columns, types.DatumRow{types.NewBytesDatum([]byte("foo"))})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "foo")

	columns[0].Type = mysql.TypeVarchar
	bs, err = dumpTextRow(nil, columns, types.DatumRow{types.NewStringDatum("bar")})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "bar")

	var d types.Datum

	time, err := types.ParseTime(nil, "2017-01-05 23:59:59.575601", mysql.TypeDatetime, 0)
	c.Assert(err, IsNil)
	d.SetMysqlTime(time)
	columns[0].Type = mysql.TypeDatetime
	bs, err = dumpTextRow(nil, columns, types.DatumRow{d})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "2017-01-06 00:00:00")

	duration, err := types.ParseDuration("11:30:45", 0)
	c.Assert(err, IsNil)
	d.SetMysqlDuration(duration)
	columns[0].Type = mysql.TypeDuration
	bs, err = dumpTextRow(nil, columns, types.DatumRow{d})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "11:30:45")

	d.SetMysqlDecimal(types.NewDecFromStringForTest("1.23"))
	columns[0].Type = mysql.TypeNewDecimal
	bs, err = dumpTextRow(nil, columns, types.DatumRow{d})
	c.Assert(err, IsNil)
	c.Assert(mustDecodeStr(c, bs), Equals, "1.23")
}

func mustDecodeStr(c *C, b []byte) string {
	str, _, _, err := parseLengthEncodedBytes(b)
	c.Assert(err, IsNil)
	return string(str)
}

func (s *testUtilSuite) TestAppendFormatFloat(c *C) {
	tests := []struct {
		fVal    float64
		out     string
		prec    int
		bitSize int
	}{
		{
			99999999999999999999,
			"1e20",
			-1,
			64,
		},
		{
			1e15,
			"1e15",
			-1,
			64,
		},
		{
			9e14,
			"900000000000000",
			-1,
			64,
		},
		{
			-9999999999999999,
			"-1e16",
			-1,
			64,
		},
		{
			999999999999999,
			"999999999999999",
			-1,
			64,
		},
		{
			0.000000000000001,
			"0.000000000000001",
			-1,
			64,
		},
		{
			0.0000000000000009,
			"9e-16",
			-1,
			64,
		},
		{
			-0.0000000000000009,
			"-9e-16",
			-1,
			64,
		},
		{
			0.11111,
			"0.111",
			3,
			64,
		},
		{
			0.11111,
			"0.111",
			3,
			64,
		},
		{
			0.1111111111111111111,
			"0.11111111",
			-1,
			32,
		},
		{
			0.1111111111111111111,
			"0.1111111111111111",
			-1,
			64,
		},
		{
			0.0000000000000009,
			"0.000",
			3,
			64,
		},
		{
			0,
			"0",
			-1,
			64,
		},
	}
	for _, t := range tests {
		c.Assert(string(appendFormatFloat(nil, t.fVal, t.prec, t.bitSize)), Equals, t.out)
	}
}
