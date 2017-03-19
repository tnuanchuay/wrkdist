package wrkdist

import (
	"strings"
	"strconv"
	"errors"
)

const (
	us	=	1e-6
	ms	=	1e-3
	s	=	1
	m	=	60
	h	=	3600
)
var listUnits = []string{"us", "ms", "s", "m", "h"}
var units = map[string]float64{
	"us":us,
	"ms":ms,
	"s":s,
	"m":m,
	"h":h,
}

func TimeToFloat(s string) (float64, error){
	var result float64 = 0
	for _, unit := range listUnits{
		if strings.Contains(s, unit){
			value := strings.Split(s, unit)
			result, err := strconv.ParseFloat(value[0], 64)
			if err != nil {
				return result, error(err)
			}
			result = result * units[unit]
			return result, nil
		}
	}
	return result, errors.New("Invalid input")
}

const (
	K = 1000.0
	M = 1000000.0
	G = 1000000000.0
)

var listSiUnits = []string{"K", "M", "G", "k", "m", "g"}
var Siunits = map[string]float64{
	"K": K,
	"M": M,
	"G": G,
	"k": K,
	"m": M,
	"g": G,
}

func SIToFloat(s string) (float64, error) {
	var result float64
	ss := string(s)
	for _, unit := range listSiUnits {
		if strings.Contains(ss, unit) {
			v := strings.Split(ss, unit)[0]
			vv, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return 0, err
			} else {
				result = float64(vv) * Siunits[unit]
				return result, nil
			}
		}
	}

	result, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return result, nil
}


