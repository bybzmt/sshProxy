
export default {
    map: {
        "/": {
            page: () => import('./pages/index.svelte'),
        },
        "/state": {
            page: () => import('./pages/state.svelte'),
        },
        "/config": {
            page: () => import('./pages/config.svelte'),
        },
        "/rules": {
            page: () => import('./pages/rules.svelte'),
        },
        "/server": {
            page: () => import('./pages/server.svelte'),
        },
    }
}
