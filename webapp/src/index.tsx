import {getConfig} from 'mattermost-redux/selectors/entities/general';
import {GlobalState} from 'mattermost-redux/types/store';
import PowerModeInput from 'power-mode-input';
import {Action, Store} from 'redux';

import manifest, {id as pluginId} from './manifest';
// eslint-disable-next-line import/no-unresolved
import {PluginRegistry} from './types/mattermost-webapp';

export const getPluginServerRoute = (state: GlobalState) => {
    const config = getConfig(state);

    let basePath = '/';

    if (config && config.SiteURL) {
        basePath = new URL(config.SiteURL).pathname;

        if (basePath && basePath[basePath.length - 1] === '/') {
            basePath = basePath.substr(0, basePath.length - 1);
        }
    }

    return basePath + '/plugins/' + pluginId;
};

function getUserConfig(state: GlobalState) {
    return new Promise((resolve, reject) => fetch(getPluginServerRoute(state)).then((r) => r.json()).then(resolve).catch(reject));
}

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        let postEl: any;
        let replyEl: any;
        let enabled = false;
        let config = {};
        let configChanged = false;
        const handleConfigChange = (c: any) => {
            enabled = c.toggle;
            config = JSON.parse(c.config);
            configChanged = true;
        };
        registry.registerWebSocketEventHandler(
            'custom_' + pluginId + '_config_changed',
            (message) => handleConfigChange(message.data),

        );
        getUserConfig(store.getState()).then(handleConfigChange);
        setInterval(() => {
            if (enabled) {
                const post = document.getElementById('post_textbox');
                const reply = document.getElementById('reply_textbox');

                if (post && (post !== postEl || configChanged)) {
                    closeSafe(postEl);
                    postEl = post;
                    PowerModeInput.make(post, config);
                }
                if (reply && (reply !== replyEl || configChanged)) {
                    closeSafe(replyEl);
                    replyEl = reply;
                    PowerModeInput.make(reply, config);
                }
            } else {
                closeSafe(replyEl);
                closeSafe(postEl);
            }
            configChanged = false;
        }, 1000);
    }
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
function closeSafe(postEl: any) {
    if (postEl) {
        try {
            PowerModeInput.close(postEl);
        } catch {
            // ignore
        }
    }
}

