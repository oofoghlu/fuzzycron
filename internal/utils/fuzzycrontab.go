package utils

import (
	"hash/fnv"
	"strconv"
	"strings"

	"github.com/robfig/cron"
)

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func moduloHash(hashNumber uint32, modulo int) string {
	return strconv.FormatUint(uint64(hashNumber%uint32(modulo)), 10)
}

func parseSchedule(schedule string) (string, error) {
	_, err := cron.ParseStandard(schedule)
	if err != nil {
		return "", err
	}
	return schedule, nil
}

func EvalCrontab(crontab string, name string) (string, error) {
	split := strings.Split(crontab, " ")
	if len(split) != 5 {
		return parseSchedule(crontab)
	}
	hashNumber := hash(name)
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
	return parseSchedule(evalSchedule)
}
