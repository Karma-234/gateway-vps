package iso

import (
	"sort"

	"github.com/moov-io/iso8583"
	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/field"
	"github.com/moov-io/iso8583/padding"
	"github.com/moov-io/iso8583/prefix"
)

var Spec8353 = &iso8583.MessageSpec{
	Name: "1987 version of ISO 8583",
	Fields: map[int]field.Field{
		0: field.NewString(&field.Spec{
			Length:      4,
			Description: "Message Type Indicator",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		1: field.NewBitmap(&field.Spec{
			Length:      16,
			Description: "Bitmap",
			Enc:         encoding.BytesToASCIIHex,
			Pref:        prefix.Hex.Fixed,
		}),

		// Key fields for card transactions
		2: field.NewString(&field.Spec{ // PAN
			Length:      19,
			Description: "Primary Account Number",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
		}),
		3: field.NewString(&field.Spec{ // Processing Code
			Length:      6,
			Description: "Processing Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		4: field.NewNumeric(&field.Spec{ // Amount
			Length:      12,
			Description: "Transaction Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
		}),
		7: field.NewString(&field.Spec{
			Length:      10,
			Description: "Transmission Date & Time",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		11: field.NewString(&field.Spec{
			Length:      6,
			Description: "System Trace Audit Number (STAN)",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		14: field.NewString(&field.Spec{
			Length:      4,
			Description: "Expiration Date",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		37: field.NewString(&field.Spec{
			Length:      12,
			Description: "Retrieval Reference Number",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		39: field.NewString(&field.Spec{
			Length:      2,
			Description: "Response Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		41: field.NewString(&field.Spec{
			Length:      8,
			Description: "Terminal ID",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		42: field.NewString(&field.Spec{
			Length:      15,
			Description: "Merchant ID",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		43: field.NewComposite(&field.Spec{ // Merchant info
			Length:      40,
			Description: "Card Acceptor Name/Location",
			Pref:        prefix.ASCII.Fixed,
			Tag: &field.TagSpec{
				Length: 2,
				Enc:    encoding.ASCII,
				Sort:   sort.Strings,
			},
			Subfields: map[string]field.Field{
				"01": field.NewString(&field.Spec{
					Length:      25,
					Description: "Card Acceptor Name",
					Enc:         encoding.ASCII,
					Pref:        prefix.ASCII.Fixed,
				}),
				"02": field.NewString(&field.Spec{
					Length:      13,
					Description: "Card Acceptor City",
					Enc:         encoding.ASCII,
					Pref:        prefix.ASCII.Fixed,
				}),
				"03": field.NewString(&field.Spec{
					Length:      2,
					Description: "Card Acceptor Country Code",
					Enc:         encoding.ASCII,
					Pref:        prefix.ASCII.Fixed,
				}),
			},
		}),
		49: field.NewString(&field.Spec{
			Length:      3,
			Description: "Currency Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
	},
}
