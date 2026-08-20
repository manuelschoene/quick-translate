import type { ViewName } from '@data/types';
import ErrorView from '@views/ErrorView.vue';
import LanguageSelectionView from '@views/LanguageSelectionView.vue';
import TranslationView from '@views/TranslationView.vue';
import type { Component } from 'vue';

/**
 * The one table that says what a view name stands for. It lives outside of `src/data`, because the
 * data layer works with the name alone and stays free of components that way.
 *
 * Nothing but the view itself is listed here. Each view brings the layout it sits in along by
 * importing it, so the frame a view needs is decided in the view and not in this table.
 *
 * The record over `ViewName` makes a view without an entry a compiler error.
 */
export const router: Record<ViewName, Component> = {
    translation: TranslationView,
    error: ErrorView,
    languageSelection: LanguageSelectionView,
};
