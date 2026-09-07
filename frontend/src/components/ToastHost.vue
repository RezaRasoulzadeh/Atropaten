<script setup lang="ts">
import { CheckCircle2, CircleX, Info, TriangleAlert, X } from 'lucide-vue-next'
import { dismiss, useToast, type ToastKind } from '../ui/feedback'
const toastItems = useToast().items
const icons: Record<ToastKind, typeof CheckCircle2> = { success: CheckCircle2, error: CircleX, warning: TriangleAlert, info: Info }
const iconFor = (kind: ToastKind) => icons[kind]
</script>
<template><div class="at-toast-host" aria-live="polite" aria-atomic="false"><TransitionGroup name="toast"><div v-for="item in toastItems" :key="item.id" class="at-toast" :class="`at-toast-${item.kind}`" :role="item.kind === 'error' ? 'alert' : 'status'"><component :is="iconFor(item.kind)" :size="17" aria-hidden="true" /><div class="at-toast-copy"><strong v-if="item.title">{{ item.title }}</strong><span>{{ item.message }}</span></div><button class="at-toast-close" type="button" aria-label="Dismiss notification" @click="dismiss(item.id)"><X :size="15" /></button></div></TransitionGroup></div></template>
