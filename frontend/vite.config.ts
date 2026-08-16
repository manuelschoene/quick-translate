import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import { defineConfig } from 'vite';

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [vue(), tailwindcss()],
    resolve: {
        alias: {
            '@': '/src',
            '@assets': '/src/assets',
            '@comp': '/src/components',
            '@data': '/src/data',
            '@lay': '/src/layouts',
            '@use': '/src/composables',
            '@utils': '/src/utils',
            '@views': '/src/views',
            '@wails': '/wailsjs',
        },
    },
});
