import { viewState } from '@data/state';

/**
 * Shows the translation. This is where the back button of the error view leads and where every
 * event that carries a translation ends up.
 */
export function showTranslation(): void {
    viewState.current = 'translation';
}

/**
 * Shows the error that was reported last. Called by `report.ts`, which is the one place a failure
 * arrives at, so nothing else has to think about bringing it up.
 */
export function showError(): void {
    viewState.current = 'error';
}

/**
 * Shows the language selection for the source language. Called by the language selector in the
 * translation view.
 */
export function showSourceSelection(): void {
    viewState.current = 'sourceSelection';
}

/**
 * Shows the language selection for the target language. Called by the language selector in the
 * translation view.
 */
export function showTargetSelection(): void {
    viewState.current = 'targetSelection';
}
