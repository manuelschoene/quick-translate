import { route } from '@/router';
import { report } from '@services/report';
import { beginShortcut, endShortcut } from '@services/request';
import { applyFull, applyTranslation } from '@services/wire';
import { Events } from '@wailsio/runtime';

/**
 * The events the backend sends on its own, mirrored from `internal/transport/adapter.go` and only to
 * be changed together with it. They are the only way a translation that was started by the shortcut
 * or by a second start of the application reaches the frontend, because those begin outside of it.
 */
const eventTranslating = 'translating';
const eventTranslation = 'translation';
const eventError = 'error';
const eventRestored = 'restored';

/**
 * Starts listening to the backend. Done before the first load, so a shortcut that is pressed
 * meanwhile is not lost.
 */
export function listen(): void {
    Events.On(eventTranslating, onTranslating);
    Events.On(eventTranslation, onTranslation);
    Events.On(eventError, onError);
    Events.On(eventRestored, onRestored);
}

/**
 * Stops listening. Never needed while the application runs, but it is in `wails dev`, where the
 * frontend is mounted again and the handlers would otherwise pile up on every event.
 */
export function silence(): void {
    Events.Off(eventTranslating, eventTranslation, eventError, eventRestored);
}

/**
 * Takes the news that the shortcut opened the window and a translation is on its way, and brings the
 * user back from an error that was left standing so the answer arrives where it is expected.
 *
 * Nothing is cleared: a translation that fails leaves the previous one to return to.
 */
function onTranslating(): void {
    beginShortcut();
    route('translation');
}

/**
 * Takes the translation the shortcut asked for. Shows the view although `onTranslating` did so
 * already, because the events do not know of each other and whoever brings the data makes sure it
 * can be seen.
 */
function onTranslation(event: Events.WailsEvent<typeof eventTranslation>): void {
    endShortcut();
    applyTranslation(event.data);
    route('translation');
}

/**
 * Takes a translation by shortcut that failed. Reporting it is what brings up the error view.
 */
function onError(event: Events.WailsEvent<typeof eventError>): void {
    endShortcut();
    report(event.data);
}

/**
 * Takes the translation a second start of the application brought back. It carries everything,
 * because a stored translation brings its own provider and languages with it.
 */
function onRestored(event: Events.WailsEvent<typeof eventRestored>): void {
    applyFull(event.data);
    route('translation');
}
