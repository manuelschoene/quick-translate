import { route } from '@/router';
import { report } from '@services/report';
import { beginShortcut, endShortcut } from '@services/request';
import { applyTranslation } from '@services/wire';
import type { transport } from '@wails/go/models';
import { EventsOff, EventsOn } from '@wails/runtime/runtime';

/**
 * The events the backend sends on its own, mirrored from `internal/transport/adapter.go` and only to
 * be changed together with it. They are the only way a translation that was started by the shortcut
 * reaches the frontend, because that one begins outside of it.
 */
const eventTranslating = 'translating';
const eventTranslation = 'translation';
const eventError = 'error';

/**
 * Starts listening to the backend. Done before the first load, so a shortcut that is pressed
 * meanwhile is not lost.
 */
export function listen(): void {
    EventsOn(eventTranslating, onTranslating);
    EventsOn(eventTranslation, onTranslation);
    EventsOn(eventError, onError);
}

/**
 * Stops listening. Never needed while the application runs, but it is in `wails dev`, where the
 * frontend is mounted again and the handlers would otherwise pile up on every event.
 */
export function silence(): void {
    EventsOff(eventTranslating, eventTranslation, eventError);
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
function onTranslation(dto: transport.TranslationDto): void {
    endShortcut();
    applyTranslation(dto);
    route('translation');
}

/**
 * Takes a translation by shortcut that failed. Reporting it is what brings up the error view.
 */
function onError(message: string): void {
    endShortcut();
    report(message);
}
