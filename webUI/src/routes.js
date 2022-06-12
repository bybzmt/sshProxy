
export default {
    map: {
        "/": {
            page: () => import('./pages/index.svelte'),
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
