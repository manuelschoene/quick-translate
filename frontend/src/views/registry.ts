import type { ViewName } from '@data/types';
import MainLayout from '@lay/MainLayout.vue';
import ErrorView from '@views/ErrorView.vue';
import TranslationView from '@views/TranslationView.vue';
import type { Component } from 'vue';

/**
 * A view together with the layout it is wrapped in.
 */
interface Entry {
    view: Component;
    layout: Component;
}

/**
 * The one table that says what a view name stands for. It lives outside of `src/data`, because the
 * data layer works with the name alone and stays free of components that way.
 *
 * The record over `ViewName` makes a view without an entry a compiler error.
 */
export const registry: Record<ViewName, Entry> = {
    translation: { view: TranslationView, layout: MainLayout },
    error: { view: ErrorView, layout: MainLayout },
};
