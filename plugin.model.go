package marionette

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func ModelSleepPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	if len(args) < 1 {
		return false, errors.New("model.sleep: not enough arguments")
	}
	distStr, ok := args[0].(string)
	if !ok {
		return false, errors.New("model.sleep: invalid argument type")
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
			return false, err
		}

		prob, err := strconv.ParseFloat(a[1], 64)
		if err != nil {
			return false, err
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

	time.Sleep(time.Duration(k * float64(time.Second)))

	return true, nil
}

func ModelSpawnPlugin(fsm *FSM, args []interface{}) (success bool, err error) {
	panic("TODO")
}

/*
# maybe these should be in marionette_state?
client_driver_ = None
server_driver_ = None
def spawn(channel, marionette_state, input_args, blocking=True):
    global client_driver_, server_driver_

    success = False

    format_name_ = input_args[0]
    num_models = int(input_args[1])

    if marionette_state.get_local("party") == 'server':
        if not server_driver_:
            driver = marionette_tg.driver.ServerDriver(
                marionette_state.get_local("party"))

            driver.set_multiplexer_incoming(
                marionette_state.get_global("multiplexer_incoming"))
            driver.set_multiplexer_outgoing(
                marionette_state.get_global("multiplexer_outgoing"))
            driver.setFormat(format_name_)
            driver.set_state(marionette_state)

            server_driver_ = driver

        if server_driver_.num_executables_completed_ < num_models:
            server_driver_.execute(reactor)
        else:
            server_driver_.stop()
            server_driver_ = None
            success = True

    elif marionette_state.get_local("party") == 'client':
        if not client_driver_:
            driver = marionette_tg.driver.ClientDriver(
                marionette_state.get_local("party"))

            driver.set_multiplexer_incoming(
                marionette_state.get_global("multiplexer_incoming"))
            driver.set_multiplexer_outgoing(
                marionette_state.get_global("multiplexer_outgoing"))
            driver.setFormat(format_name_)
            driver.set_state(marionette_state)

            driver.reset(num_models)
            client_driver_ = driver

        if client_driver_.isRunning():
            client_driver_.execute(reactor)
        else:
            client_driver_.stop()
            client_driver_ = None
            success = True

    return success
*/
