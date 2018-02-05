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
		"bind": channelBindPlugin,
	},
	"fte": map[string]PluginFunc{
		"send_async": fteSendAsyncPlugin,
		"recv_async": fteRecvAsyncPlugin,
		"send":       fteSendSyncPlugin,
		"recv":       fteRecvSyncPlugin,
	},
	"io": map[string]PluginFunc{
		"puts": ioPutsPlugin,
		"gets": ioGetsPlugin,
	},
	"model": map[string]PluginFunc{
		"sleep": modelSleepPlugin,
		"spawn": modelSpawnPlugin,
	},
	"tg": map[string]PluginFunc{
		"send": tgSendPlugin,
		"recv": tgRecvPlugin,
	},
}
