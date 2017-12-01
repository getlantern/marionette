package marionette

type PluginFunc func(fsm *FSM, args []interface{}) (success bool, err error)

// FindPlugin returns a plugin function by module & name.
func FindPlugin(module, method string) PluginFunc {
	m := pluginRegistry[module]
	if m == nil {
		return nil
	}
	return m[method]
}

var pluginRegistry = map[string]map[string]PluginFunc{
	"channel": map[string]PluginFunc{
		"bind": ChannelBindPlugin,
	},
	"fte": map[string]PluginFunc{
		"send_async": FTESendAsyncPlugin,
		"recv_async": FTERecvAsyncPlugin,
		"send":       FTESendPlugin,
		"recv":       FTERecvPlugin,
	},
	"io": map[string]PluginFunc{
		"puts": IOPutsPlugin,
		"gets": IOGetsPlugin,
	},
	"model": map[string]PluginFunc{
		"sleep": ModelSleepPlugin,
		"spawn": ModelSpawnPlugin,
	},
	"tg": map[string]PluginFunc{
		"send": TGSendPlugin,
	},
}
