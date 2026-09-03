import { historyState } from '@/state';
import { readonlyRefs } from '@lib/reactivity';
import { request } from '@services/request';
import { applyFull } from '@services/wire';
import { NextTranslation, PreviousTranslation } from '@wails/go/transport/Adapter';

const history = readonlyRefs(historyState);

const exposed = { ...history, previousTranslation, nextTranslation };

/**
 * Gives the components the ways to step through the translations that were made before and whether
 * there is anything to step to. Both flags stay false while the history is turned off.
 *
 * A stored translation is shown as it was made and is not translated again. It brings its provider
 * and its languages with it, which the other composables pick up on their own.
 */
export function useHistory(): typeof exposed {
    return exposed;
}

/**
 * Steps to the translation that was stored before the current one.
 */
async function previousTranslation(): Promise<void> {
    await request(PreviousTranslation, applyFull);
}

/**
 * Steps to the translation that was stored after the current one.
 */
async function nextTranslation(): Promise<void> {
    await request(NextTranslation, applyFull);
}
