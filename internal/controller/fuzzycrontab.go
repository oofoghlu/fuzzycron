package controller

import (
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/robfig/cron"
)

func IsValidCrontab(Crontab string) bool {
	return true
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func moduloHash(hashNumber uint32, modulo int) string {
	return strconv.FormatUint(uint64(hashNumber%uint32(modulo)), 10)
}

func EvalCrontab(Crontab string, Name string) (string, error) {
	split := strings.Split(Crontab, " ")
	if len(split) != 5 {
		return Name, fmt.Errorf("%s crontab must contain exactly 5 elements", Name)
	}
	hashNumber := hash(Name)
	var evalSplit [5]string
	for index, num := range split {
		switch index {
		case 0:
			// minute
			if num == "H" {
				evalSplit[index] = moduloHash(hashNumber, 60)
			} else {
				evalSplit[index] = num
			}
		case 1:
			// hour
			if num == "H" {
				evalSplit[index] = moduloHash(hashNumber, 24)
			} else {
				evalSplit[index] = num
			}
		default:
			// do nothing for now
			evalSplit[index] = num
		}
	}
	evalSchedule := strings.Join(evalSplit[:], " ")
	_, err := cron.ParseStandard(evalSchedule)
	if err != nil {
		return "", err
	}
	return evalSchedule, nil
}
