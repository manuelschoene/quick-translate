import { historyState } from '@/state';
import { NextTranslation, PreviousTranslation } from '@bind/transport/adapter';
import { readonlyRefs } from '@lib/reactivity';
import { request } from '@services/request';
import { applyFull } from '@services/wire';

const history = readonlyRefs(historyState);

const exposed = { ...history, previousTranslation, nextTranslation };

/**
 * Gives the components the ways to step through the translations that were made before and whether
 * there is anything to step to.
 */
export function useHistory(): typeof exposed {
    return exposed;
}

/**
 * Steps to the translation that was stored before the current one. It is shown as it was made and
 * not translated again, and it brings its own provider and languages along.
 */
async function previousTranslation(): Promise<void> {
    await request(PreviousTranslation, applyFull);
}

/**
 * Steps to the translation that was stored after the current one, the same way back.
 */
async function nextTranslation(): Promise<void> {
    await request(NextTranslation, applyFull);
}
