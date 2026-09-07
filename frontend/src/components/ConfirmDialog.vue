<script setup lang="ts">
import { confirmState, resolveConfirm } from '../ui/feedback'
</script>
<template><div v-if="confirmState.open" class="modal modal-open" role="dialog" aria-modal="true" @keydown.esc="resolveConfirm(false)"><div class="modal-box"><h3 class="text-lg font-bold">{{ confirmState.title }}</h3><p class="text-sm text-base-content/70">{{ confirmState.message }}</p><div class="modal-action"><button class="btn btn-ghost" type="button" @click="resolveConfirm(false)">{{ confirmState.cancelLabel }}</button><button class="btn" :class="confirmState.danger ? 'btn-error' : 'btn-primary'" type="button" @click="resolveConfirm(true)">{{ confirmState.confirmLabel }}</button></div></div><div class="modal-backdrop" @click="resolveConfirm(false)"></div></div></template>
