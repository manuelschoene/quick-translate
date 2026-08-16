import { applyFull, applyLanguages, applyProviders, applyTranslation } from '@data/apply';
import { listen, silence } from '@data/events';
import { request } from '@data/request';
import {
    ChangeProvider,
    ChangeSource,
    ChangeTarget,
    CopyTranslation,
    Hide,
    Languages,
    NextTranslation,
    PreviousTranslation,
    Providers,
    SwitchLanguages,
    Translation,
} from '@wails/go/transport/Adapter';

/**
 * Brings the application up: listen to the events first, then fill it with the current state.
 */
export async function start(): Promise<void> {
    listen();
    await load();
}

/**
 * Takes it down again, as far as there is anything to take down in the frontend.
 */
export function stop(): void {
    silence();
}

/**
 * Loads the whole state of the application. Meant for the first render and for a frontend that was
 * reloaded while a translation was on screen. Nothing is translated by this, the language lists come
 * from the cache of the backend.
 */
export async function load(): Promise<void> {
    await Promise.all([
        request(Providers, applyProviders),
        request(Languages, applyLanguages),
        request(Translation, applyTranslation),
    ]);
}

/**
 * Switches to the given provider and translates the current text with it.
 */
export async function changeProvider(slug: string): Promise<void> {
    await request(() => ChangeProvider(slug), applyFull);
}

/**
 * Sets the source language and translates the current text again.
 */
export async function changeSource(tag: string): Promise<void> {
    await request(() => ChangeSource(tag), applyTranslation);
}

/**
 * Sets the target language and translates the current text again.
 */
export async function changeTarget(tag: string): Promise<void> {
    await request(() => ChangeTarget(tag), applyTranslation);
}

/**
 * Switches the source and the target language and translates the current text in the other
 * direction.
 */
export async function switchLanguages(): Promise<void> {
    await request(SwitchLanguages, applyTranslation);
}

/**
 * Steps to the translation that was stored before the current one. Nothing is translated again, the
 * stored translation is shown as it was made.
 */
export async function previousTranslation(): Promise<void> {
    await request(PreviousTranslation, applyFull);
}

/**
 * Steps to the translation that was stored after the current one. Nothing is translated again, the
 * stored translation is shown as it was made.
 */
export async function nextTranslation(): Promise<void> {
    await request(NextTranslation, applyFull);
}

/**
 * Writes the current translation to the clipboard. The window hides itself when this succeeds, as
 * the translation is on its way into another application at that point.
 */
export async function copyTranslation(): Promise<void> {
    await request(CopyTranslation);
}

/**
 * Hides the window without stopping the application, which is what dismissing it does.
 */
export async function hide(): Promise<void> {
    await request(Hide);
}
