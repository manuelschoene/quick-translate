// Imported for its side effects: the runtime is what makes the frameless window draggable through
// `--wails-draggable` and what applies `--default-contextmenu`. Importing it here keeps both working
// regardless of which other module happens to import something from it.
import '@wailsio/runtime';
import { createApp } from 'vue';
import App from './App.vue';
import './style.css';

createApp(App).mount('#app');
