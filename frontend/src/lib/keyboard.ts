import { onMounted, onUnmounted } from 'vue';

/**
 * Runs the action whenever the key is pressed, for as long as the view that asks for it is on screen.
 *
 * Meant for the keys that belong to a whole view, `Escape` above all: the listener sits on the window
 * so that it fires while the search box holds the focus. A key that belongs to a single element is
 * better handled on that element.
 *
 * A key that only means something in part of the view takes a guard. While it answers false the key
 * is left to the browser untouched, which is what keeps the arrow keys moving the caret as long as
 * the user is still typing.
 */
export function onKey(key: string, action: () => void, when?: () => boolean): void {
    function handle(event: KeyboardEvent): void {
        if (event.key !== key) return;
        if (when && !when()) return;

        event.preventDefault();
        action();
    }

    onMounted(() => {
        window.addEventListener('keydown', handle);
    });
    onUnmounted(() => {
        window.removeEventListener('keydown', handle);
    });
}
