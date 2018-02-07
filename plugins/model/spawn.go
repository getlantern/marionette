package marionette

import (
	"github.com/redjack/marionette"
)

func init() {
	marionette.RegisterPlugin("model", "spawn", Spawn)
}

func Spawn(fsm *marionette.FSM, args []interface{}) (success bool, err error) {
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
