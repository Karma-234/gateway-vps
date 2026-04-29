package iso

import (
	"github.com/moov-io/iso8583"
	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/field"
	"github.com/moov-io/iso8583/prefix"
)

var spec8353 = &iso8583.MessageSpec{
	Name: "1987 version of ISO 8583",
	Fields: map[int]field.Field{
		0: field.NewString(&field.Spec{Length: 4, Description: "MTI", Enc: encoding.ASCII, Pref: prefix.ASCII.Fixed}),
		1: field.NewBitmap(&field.Spec{Length: 16, Description: "Bitmap", Enc: encoding.BytesToASCIIHex, Pref: prefix.Hex.Fixed}),
	},
}
