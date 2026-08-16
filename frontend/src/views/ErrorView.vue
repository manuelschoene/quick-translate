<!--
    A placeholder that shows the error that was reported last, with the one way out of it. The
    button is the case where a component asks for the view to change.
-->
<script lang="ts" setup>
import { useError } from '@use/useError';

const { message, showTranslation } = useError();
</script>

<template>
    <p>{{ message }}</p>
    <button @click="showTranslation" type="button">Back</button>
</template>
