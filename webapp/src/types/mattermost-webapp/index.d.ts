export interface PluginRegistry {
    registerWebSocketEventHandler(id: string, handler: (data) => void)
    registerPostTypeComponent(typeName: string, component: React.ElementType)

    // Add more if needed from https://developers.mattermost.com/extend/plugins/webapp/reference
}
