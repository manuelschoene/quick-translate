import { onMounted, onUnmounted } from 'vue';

/**
 * Runs the action whenever the key is pressed, for as long as the view that asks for it is on screen.
 *
 * Meant for the keys that belong to a whole view, `Escape` above all: the listener sits on the window
 * so that it fires while the search box holds the focus. A key that belongs to a single element is
 * better handled on that element.
 */
export function onKey(key: string, action: () => void): void {
    function handle(event: KeyboardEvent): void {
        if (event.key !== key) return;

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
