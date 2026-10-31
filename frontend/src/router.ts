import ErrorView from '@views/ErrorView.vue';
import SelectionSourceView from '@views/SelectionSourceView.vue';
import SelectionTargetView from '@views/SelectionTargetView.vue';
import TranslationView from '@views/TranslationView.vue';
import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router';

/**
 * Every view by name. The paths live in the hash and never reach the asset server, so they cannot collide with
 * the `/wails/*` paths Wails serves.
 */
const views = [
    { name: 'translation', path: '/', component: TranslationView },
    { name: 'error', path: '/error', component: ErrorView },
    { name: 'sourceSelection', path: '/source', component: SelectionSourceView },
    { name: 'targetSelection', path: '/target', component: SelectionTargetView },
] as const satisfies readonly RouteRecordRaw[];

export type ViewName = (typeof views)[number]['name'];

/**
 * The router of the window. Hash history, because a real path would be requested from the asset server, which
 * neither falls back to `index.html` nor leaves `/wails/*` to the frontend. An unknown path leads to the
 * translation.
 */
export const router = createRouter({
    history: createWebHashHistory(),
    routes: [...views, { path: '/:pathMatch(.*)*', redirect: '/' }],
});

/**
 * Takes the user to the given view. Usable outside of components, such as from the event handlers. Going to the
 * view already shown does nothing.
 */
export function route(name: ViewName): void {
    void router.push({ name });
}
