import { listen, silence } from '@services/events';
import { request, requestQuietly } from '@services/request';
import { applyFull } from '@services/wire';
import { Hide, State } from '@wails/go/transport/Adapter';

const exposed = { start, stop, hide };

/**
 * Gives the components what belongs to the application as a whole: the ways to bring it up and take
 * it down again and the way to dismiss it.
 */
export function useApplication(): typeof exposed {
    return exposed;
}

/**
 * Brings the application up: listen to the events first, then fill it with the current state.
 */
async function start(): Promise<void> {
    listen();
    await load();
}

/**
 * Takes it down again, as far as there is anything to take down in the frontend.
 */
function stop(): void {
    silence();
}

/**
 * Fetches the whole state. Meant for the first render and for a frontend that was reloaded while a
 * translation was on screen. Nothing is translated by it, the language lists come from the cache of
 * the backend.
 */
async function load(): Promise<void> {
    await request(State, applyFull);
}

/**
 * Hides the window without stopping the application. Asked for quietly: a window that refuses to go
 * away is worth a message on the console, but not the error view in the window that is still there.
 */
async function hide(): Promise<void> {
    await requestQuietly(Hide);
}
