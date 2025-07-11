package parsers

import (
	"errors"
	"io"
	"runtime"
)

func ParseInt(s string) (int, error) {
	n := 0
	if len(s) == 0 {
		return 0, errors.New("empty string cannot be parsed to int")
	}
	sign := 1
	if s[0] == '-' {
		sign = -1
		s = s[1:]
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return n * sign, errors.New("invalid character in string, must be a digit")
		}
		n = n*10 + int(r-'0')
	}
	n *= sign
	return n, nil
}

func ParseFloat(s string) (float64, error) {
	if len(s) == 0 {
		return 0, errors.New("empty string cannot be parsed to float")
	}
	n := 0.0
	sign := 1.0
	if s[0] == '-' {
		sign = -1.0
		s = s[1:]
	}
	decimalPoint := false
	factor := 1.0
	for _, r := range s {
		if r == '.' {
			if decimalPoint {
				return n * sign, errors.New("multiple decimal points in string")
			}
			decimalPoint = true
			continue
		}
		if r < '0' || r > '9' {
			return n * sign, errors.New("invalid character in string, must be a digit or decimal point")
		}
		if decimalPoint {
			factor *= 10.0
			n += float64(r-'0') / factor
		} else {
			n = n*10 + float64(r-'0')
		}
	}
	n *= sign
	return n, nil
}

func ParseBool(s string) (bool, error) {
	switch s {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, errors.New("invalid boolean string, must be true/false, 1/0, yes/no, on/off")
	}
}

func ReadLine(reader io.Reader) ([]byte, error) {
	var buf [1]byte
	var ret []byte

	for {
		n, err := reader.Read(buf[:])
		if n > 0 {
			switch buf[0] {
			case '\b':
				if len(ret) > 0 {
					ret = ret[:len(ret)-1]
				}
			case '\n':
				if runtime.GOOS != "windows" {
					return ret, nil
				}
				// otherwise ignore \n
			case '\r':
				if runtime.GOOS == "windows" {
					return ret, nil
				}
				// otherwise ignore \r
			default:
				ret = append(ret, buf[0])
			}
			continue
		}
		if err != nil {
			if err == io.EOF && len(ret) > 0 {
				return ret, nil
			}
			return ret, err
		}
	}
}
