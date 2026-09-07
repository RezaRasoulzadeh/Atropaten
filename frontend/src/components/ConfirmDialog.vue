<script setup lang="ts">
import { confirmState, resolveConfirm } from '../ui/feedback'
</script>
<template><div v-if="confirmState.open" class="modal modal-open at-modal" role="dialog" aria-modal="true" @keydown.esc="resolveConfirm(false)"><div class="modal-box at-confirm-box"><h3>{{ confirmState.title }}</h3><p>{{ confirmState.message }}</p><div class="modal-action"><button class="btn at-btn at-btn-ghost" type="button" @click="resolveConfirm(false)">{{ confirmState.cancelLabel }}</button><button class="btn at-btn" :class="confirmState.danger ? 'at-btn-danger' : 'at-btn-primary'" type="button" @click="resolveConfirm(true)">{{ confirmState.confirmLabel }}</button></div></div><div class="modal-backdrop" @click="resolveConfirm(false)"></div></div></template>
