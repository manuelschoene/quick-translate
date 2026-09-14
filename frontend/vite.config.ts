import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import wails from '@wailsio/runtime/plugins/vite';
import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
    // Loads the generated event bindings into the runtime, so an event's data arrives as its DTO class.
    plugins: [vue(), tailwindcss(), wails('./bindings')],
    resolve: {
        alias: {
            '@': '/src',
            '@assets': '/src/assets',
            '@comp': '/src/components',
            '@lay': '/src/layouts',
            '@lib': '/src/lib',
            '@services': '/src/services',
            '@use': '/src/composables',
            '@views': '/src/views',
            '@bind': '/bindings/quick-translate/internal',
        },
    },
});
