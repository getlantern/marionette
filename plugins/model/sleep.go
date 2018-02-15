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
	t0 := time.Now()

	logger := marionette.Logger.With(
		zap.String("plugin", "model.sleep"),
		zap.String("party", fsm.Party()),
		zap.String("state", fsm.State()),
	)

	if len(args) < 1 {
		return errors.New("not enough arguments")
	}
	distStr, ok := args[0].(string)
	if !ok {
		return errors.New("invalid argument type")
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
			logger.Error("cannot parse element", zap.Int("i", 0), zap.String("value", a[0]), zap.Error(err))
			return err
		}

		prob, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			logger.Error("cannot parse element", zap.Int("i", 1), zap.String("value", a[1]), zap.Error(err))
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
	time.Sleep(duration)

	logger.Debug("sleep complete", zap.Duration("duration", duration), zap.Duration("t", time.Since(t0)))

	return nil
}
