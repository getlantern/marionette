package model

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/redjack/marionette"
	"go.uber.org/zap"
)

func init() {
	marionette.RegisterPlugin("model", "sleep", Sleep)
}

func Sleep(fsm marionette.FSM, args ...interface{}) error {
	logger := marionette.Logger.With(zap.String("party", fsm.Party()), zap.String("state", fsm.State()))

	if len(args) < 1 {
		return errors.New("model.sleep: not enough arguments")
	}
	distStr, ok := args[0].(string)
	if !ok {
		return errors.New("model.sleep: invalid argument type")
	}

	// Parse distribution.
	distStr = distStr[1:]
	distStr = strings.Replace(distStr, " ", "", -1)
	distStr = strings.Replace(distStr, "\n", "", -1)
	distStr = strings.Replace(distStr, "\t", "", -1)
	distStr = strings.Replace(distStr, "\r", "", -1)

	dist := make(map[float64]float64)
	var keys []float64
	for _, item := range strings.Split(distStr, ",") {
		a := strings.Split(item, ":")
		a[0] = strings.Trim(a[0], "'")

		val, err := strconv.ParseFloat(a[0], 64)
		if err != nil {
			return err
		}

		prob, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			return err
		}

		if val > 0 {
			dist[val] = prob
			keys = append(keys, val)
		}
	}

	sum, coin := float64(0), rand.Float64()
	var k float64
	for _, k = range keys {
		sum += dist[k]
		if sum >= coin {
			break
		}
	}

	duration := time.Duration(k * float64(time.Second))
	logger.Debug("model.sleep", zap.Duration("duration", duration))

	time.Sleep(duration)

	return nil
}
