package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	userId := r.Header.Get("Mattermost-User-Id")
	if userId == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	d, err := p.API.KVGet(userId)
	result := UserSettings{
		Toggle: false,
		Config: "{}",
	}
	if err == nil && d != nil {
		json.Unmarshal(d, &result)
	}
	responseJSON, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(responseJSON); err != nil {
		p.API.LogError("Failed to write status", "err", err.Error())
	}
}

func getAutocompleteData() *model.AutocompleteData {
	command := model.NewAutocompleteData("powermode", "", "Enables or disables power mode.")
	command.AddStaticListArgument("", true, []model.AutocompleteListItem{
		{
			Item:     "on",
			HelpText: "Toggle power mode on",
		}, {
			Item:     "off",
			HelpText: "Toggle power mode off",
		},
	})

	return command
}

func (p *Plugin) OnActivate() error {
	p.API.RegisterCommand(&model.Command{
		Trigger:          "powermode",
		AutoComplete:     true,
		AutoCompleteHint: "(on|off)",
		AutoCompleteDesc: "Powermode for Mattermost",
		AutocompleteData: getAutocompleteData(),
	})
	return nil
}

type UserSettings struct {
	Toggle bool   `json:"toggle"`
	Config string `json:"config"`
}

const example = `
{
  "height": 5,
  "tha": [0, 360],
  "g": 0.5,
  "num": 5,
  "radius": 6,
  "circle": true,
  "alpha": [0.75, 0.1],
  "color": "random"
}
`
const help = "###### Power mode!\n" +
	"- `/powermode on` - Set up power mode with default.\n" +
	"- `/powermode off` - Disable power mode.\n" +
	"- `/powermode on {custom config}` - Set up power mode with custom config. Here's an example: \n```" + example + "```\n" +
	"- `/powermode help` - Show this help text"

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if strings.HasPrefix(args.Command, "/powermode help") {
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         help,
		}, nil
	} else if strings.HasPrefix(args.Command, "/powermode on") {
		confStr := strings.TrimSpace(strings.Replace(args.Command, "/powermode on", "", 1))
		if confStr == "" {
			confStr = "{}"
		}
		v, _ := json.Marshal(UserSettings{Toggle: true, Config: confStr})
		p.API.KVSetWithOptions(args.UserId, v, model.PluginKVSetOptions{})
		p.API.PublishWebSocketEvent("config_changed", map[string]interface{}{"toggle": true, "config": confStr}, &model.WebsocketBroadcast{
			UserId: args.UserId,
		})
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "You are ready to rock!",
		}, nil
	} else if args.Command == "/powermode off" {
		v, _ := json.Marshal(UserSettings{Toggle: false, Config: "{}"})
		p.API.KVSetWithOptions(args.UserId, v, model.PluginKVSetOptions{})
		p.API.PublishWebSocketEvent("config_changed", map[string]interface{}{"toggle": false, "config": "{}"}, &model.WebsocketBroadcast{
			UserId: args.UserId,
		})
		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         "Sad :(",
		}, nil
	}
	return &model.CommandResponse{}, nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/
