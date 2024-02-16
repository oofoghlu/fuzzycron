package utils

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"

	"github.com/robfig/cron"
)

var upperBounds = [...]int{59, 23, 31, 12, 7}
var lowerBounds = [...]int{0, 0, 1, 1, 0}
var hashRangeRegex = regexp.MustCompile(`H\((?P<start>\d+)\-(?P<end>\d+)\)`)

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func moduloHash(hashNumber uint32, modulo int, offset int) string {
	return strconv.FormatUint(uint64((hashNumber%uint32(modulo))+uint32(offset)), 10)
}

func parseSchedule(schedule string) (string, error) {
	_, err := cron.ParseStandard(schedule)
	if err != nil {
		return "", err
	}
	return schedule, nil
}

func parseRangeField(field string, index int, hashNumber uint32, match []string) (string, error) {
	startNum, error := strconv.Atoi(match[1])
	if error != nil {
		return field, error
	}
	endNum, error := strconv.Atoi(match[2])
	if error != nil {
		return field, error
	}
	return moduloHash(hashNumber, endNum-startNum, startNum), nil
}

func parseStepField(field string, index int, hashNumber uint32) (string, error) {
	numSplit := strings.Split(field, "/")
	step, error := strconv.Atoi(numSplit[1])
	if error == nil {
		if numSplit[0] == "H" {
			return fmt.Sprintf("%s/%d", moduloHash(hashNumber, step, lowerBounds[index]), step), nil
		} else if match := hashRangeRegex.FindStringSubmatch(numSplit[0]); match != nil {
			evaluatedRange, error := parseRangeField(field, index, hashNumber, match)
			if error != nil {
				return field, error
			}
			return fmt.Sprintf("%s/%d", evaluatedRange, step), nil
		}
	}
	return field, error
}

func parseField(field string, index int, hashNumber uint32) (string, error) {
	if field == "H" {
		return moduloHash(hashNumber, upperBounds[index]+1-lowerBounds[index], lowerBounds[index]), nil
	} else if len(strings.Split(field, "/")) == 2 {
		return parseStepField(field, index, hashNumber)
	} else if match := hashRangeRegex.FindStringSubmatch(field); match != nil {
		evaluatedRange, error := parseRangeField(field, index, hashNumber, match)
		if error != nil {
			return field, error
		}
		return evaluatedRange, nil
	}
	return field, nil
}

func EvalCrontab(crontab string, name string) (string, error) {
	split := strings.Split(crontab, " ")
	if len(split) != 5 {
		return parseSchedule(crontab)
	}
	hashNumber := hash(name)
	var evalSplit [5]string
	for index, field := range split {
		evaluatedField, error := parseField(field, index, hashNumber)
		if error != nil {
			return "", error
		}
		evalSplit[index] = evaluatedField
	}
	evalSchedule := strings.Join(evalSplit[:], " ")
	return parseSchedule(evalSchedule)
}
