import { historyState } from '@/state';
import type { Navigation as NavigationModel } from '@bind/translation';
import { Navigation, Next, Previous } from '@bind/translation/service';
import { on } from '@lib/events';
import { readonlyRefs } from '@lib/reactivity';
import { request } from '@services/request';

const exposed = { ...readonlyRefs(historyState), previousTranslation, nextTranslation };

/**
 * Gives the components the ways to step through the history and whether there is anything to step to.
 */
export function useHistory(): typeof exposed {
    return exposed;
}

/**
 * Listens to the changes of the history position. Returns the unsubscribe.
 */
export function listen(): () => void {
    return on('history:changed', apply);
}

/**
 * Fetches the history position, for the first render.
 */
export async function load(): Promise<void> {
    apply(await Navigation());
}

/**
 * Steps to the translation stored before the shown one. It arrives with its provider and languages through
 * the events of those domains.
 */
async function previousTranslation(): Promise<void> {
    await request(Previous);
}

/**
 * Steps to the translation stored after the shown one.
 */
async function nextTranslation(): Promise<void> {
    await request(Next);
}

/**
 * Takes the history position over.
 */
function apply(navigation: NavigationModel): void {
    historyState.hasPrevious = navigation.hasPrevious;
    historyState.hasNext = navigation.hasNext;
}
