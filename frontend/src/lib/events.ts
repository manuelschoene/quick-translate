import { Events } from '@wailsio/runtime';

/**
 * Listens to a backend event and hands its data to the handler. Only takes names the bindings declare, because
 * `Events.On` accepts any string and silently types the data as `any`. Returns the unsubscribe.
 */
export function on<E extends keyof Events.CustomEvents>(
    name: E,
    handler: (data: Events.CustomEvents[E]) => void,
): () => void {
    return Events.On(name, (event) => {
        handler(event.data);
    });
}
