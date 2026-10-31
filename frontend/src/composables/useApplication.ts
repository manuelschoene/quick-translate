import { request, requestQuietly } from '@services/request';
import * as history from '@use/useHistory';
import * as languages from '@use/useLanguages';
import * as providers from '@use/useProviders';
import * as translation from '@use/useTranslation';
import { Window } from '@wailsio/runtime';

const domains = [providers, languages, translation, history];

/**
 * The unsubscribes of every domain, collected on start.
 */
let offs: (() => void)[] = [];

const exposed = { start, stop, hide };

/**
 * Gives the components what belongs to the application as a whole: bringing it up and down and dismissing it.
 */
export function useApplication(): typeof exposed {
    return exposed;
}

/**
 * Listens to every domain first, so a shortcut pressed meanwhile is not lost, then fetches the state. A getter
 * that fails, such as one asked before a provider is configured, shows the error view.
 */
async function start(): Promise<void> {
    offs = domains.map((domain) => domain.listen());

    await Promise.all(domains.map((domain) => request(domain.load)));
}

/**
 * Stops listening. Needed in `wails dev`, where the frontend is mounted again.
 */
function stop(): void {
    offs.forEach((off) => {
        off();
    });
    offs = [];
}

/**
 * Hides the window without stopping the application and answers whether it worked.
 */
async function hide(): Promise<boolean> {
    return await requestQuietly(() => Window.Hide());
}
