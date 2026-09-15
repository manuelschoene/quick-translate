import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import wails from '@wailsio/runtime/plugins/vite';
import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => ({
    plugins: [vue(), tailwindcss(), wails('./bindings')],
    // `wails3 dev` picks the port, exports it as WAILS_VITE_PORT and points the application at it, so the
    // dev server must take exactly that one instead of moving on to the next free port.
    server: {
        host: '127.0.0.1',
        port: Number(process.env.WAILS_VITE_PORT) || 9245,
        strictPort: true,
    },
    build: {
        rolldownOptions: {
            output: {
                // Leftover debugging calls, console output included, are stripped from production builds.
                // Replaces Vite's own minify setting, so it must not reach `build:dev`, which is built with
                // `--minify false` and keeps them.
                ...(mode === 'production' && { minify: { compress: { dropConsole: true, dropDebugger: true } } }),
                // The dependencies change far less often than the application code.
                codeSplitting: {
                    groups: [{ name: 'vendor', test: /node_modules/ }],
                },
            },
        },
    },
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
}));
