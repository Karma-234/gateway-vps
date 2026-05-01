package pkg

func MaskPAN(pan string) string {
	if len(pan) <= 10 {
		return pan
	}
	return pan[:6] + "******" + pan[len(pan)-4:]
}

func ResponseMTI(reqMTI string) string {
	switch reqMTI {
	case "0100":
		return "0110"
	case "0200":
		return "0210"
	case "0400":
		return "0410"
	default:
		// Fallback: replace last two digits with 10 if MTI format is valid
		if len(reqMTI) == 4 {
			return reqMTI[:2] + "10"
		}
		return "0210"
	}
}
