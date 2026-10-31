import { route } from '@/router';
import { translationState } from '@/state';
import type { Translation } from '@bind/models';
import { Copy, Current, Retranslate } from '@bind/translation/service';
import { on } from '@lib/events';
import { readonlyRefs } from '@lib/reactivity';
import { report } from '@services/report';
import { requestQuietly } from '@services/request';

const exposed = { ...readonlyRefs(translationState), copyTranslation };

/**
 * Gives the components the translation on screen, whether one is being made and the way to copy it.
 */
export function useTranslation(): typeof exposed {
    return exposed;
}

/**
 * Listens to the translation events. Every source of a translation (shortcut, retranslation, history step)
 * reaches the frontend through them, so they also bring the translation view up. Returns the unsubscribe.
 */
export function listen(): () => void {
    const offs = [
        on('translation:started', () => {
            translationState.pending = true;
            route('translation');
        }),
        on('translation:changed', (translation) => {
            apply(translation);
            translationState.pending = false;
            route('translation');
        }),
        // The previous translation stays, to return to from the error view.
        on('translation:failed', (message) => {
            translationState.pending = false;
            report(message);
        }),
    ];

    return () => {
        offs.forEach((off) => {
            off();
        });
    };
}

/**
 * Fetches the translation on screen, for the first render. Nothing before the first translation.
 */
export async function load(): Promise<void> {
    const current = await Current();

    if (current) {
        apply(current);
    }
}

/**
 * Translates the shown text again, for composing with a change of language or provider. Does nothing while no
 * translation is shown, because the backend has nothing to translate then.
 */
export async function retranslate(): Promise<void> {
    if (translationState.shown) {
        await Retranslate();
    }
}

/**
 * Writes the translation to the clipboard and answers whether it worked.
 */
async function copyTranslation(): Promise<boolean> {
    return await requestQuietly(Copy);
}

/**
 * Takes a translation over as the one on screen.
 */
function apply(translation: Translation): void {
    translationState.translation = translation.translation;
    translationState.shown = true;
}
