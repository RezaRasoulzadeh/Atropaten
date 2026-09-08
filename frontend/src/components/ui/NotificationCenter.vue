<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { Bell, CheckCheck, RefreshCw, X } from 'lucide-vue-next';
import { reportsApi } from '../../api/reports';
import { currentCanonicalDate, formatDateTime } from '../../utils/date';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { normalizeError, useNotifications, useToast } from '../../ui/feedback';
const props = defineProps<{ currencyUnit: CurrencyUnit; refreshKey: string }>();
const emit = defineEmits<{ navigate: [view: string] }>();
const feed = useNotifications();
const open = ref(false);
const loading = ref(false);
const toast = useToast();
const host = ref<HTMLElement | null>(null);
const trigger = ref<HTMLButtonElement | null>(null);
const unreadOnly = ref(false);
const visible = computed(() =>
  feed.items.value.filter((item) => !unreadOnly.value || !feed.isRead(item.id)),
);
let timer: ReturnType<typeof setInterval> | undefined;
async function refresh() {
  if (loading.value) return;
  loading.value = true;
  try {
    const today = currentCanonicalDate();
    feed.update(await reportsApi.dashboard(today, today));
  } catch (e) {
    toast.error(normalizeError(e).message, 'Notifications');
  } finally {
    loading.value = false;
  }
}
function close() {
  open.value = false;
  trigger.value?.focus();
}
function outside(event: MouseEvent) {
  if (!host.value?.contains(event.target as Node)) open.value = false;
}
function keydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && open.value) {
    event.stopPropagation();
    close();
  }
}
function activate(id: string, destination: string) {
  feed.markRead(id);
  close();
  emit('navigate', destination);
}
watch(() => props.refreshKey, refresh);
watch(open, (value) => {
  if (value) refresh();
});
onMounted(() => {
  refresh();
  timer = setInterval(refresh, 60000);
  window.addEventListener('focus', refresh);
  document.addEventListener('click', outside);
});
onBeforeUnmount(() => {
  clearInterval(timer);
  window.removeEventListener('focus', refresh);
  document.removeEventListener('click', outside);
});
</script>
<template>
  <div ref="host" class="relative shrink-0" @keydown="keydown">
    <button
      ref="trigger"
      class="btn btn-ghost btn-square relative"
      type="button"
      :aria-label="`Notifications, ${feed.unreadCount.value} unread`"
      :aria-expanded="open"
      aria-controls="notification-center"
      @click="open = !open"
    >
      <Bell :size="18" aria-hidden="true" />
      <span
        v-if="feed.unreadCount.value"
        class="badge badge-primary badge-xs absolute -end-1 -top-1"
        aria-hidden="true"
        >{{ feed.unreadCount.value }}</span
      >
    </button>
    <section
      v-if="open"
      id="notification-center"
      aria-label="Notifications"
      class="absolute end-0 top-full z-40 mt-2 w-96 max-w-[calc(100vw-2rem)] rounded-box border border-base-300 bg-base-100 shadow-lg"
    >
      <header class="flex items-center gap-2 border-b border-base-300 p-3">
        <h2 class="flex-1 text-sm font-semibold">Notifications</h2>
        <button
          class="btn btn-ghost btn-square btn-sm"
          :disabled="loading"
          aria-label="Refresh notifications"
          @click="refresh"
        >
          <RefreshCw :size="15" :class="{ 'animate-spin': loading }" />
        </button>
        <button
          class="btn btn-ghost btn-square btn-sm"
          aria-label="Close notifications"
          @click="close"
        >
          <X :size="16" />
        </button>
      </header>
      <div
        class="flex items-center justify-between gap-2 border-b border-base-300 px-3 py-2"
      >
        <label class="flex items-center gap-2 text-xs"
          ><input
            v-model="unreadOnly"
            type="checkbox"
            class="checkbox checkbox-xs"
          />
          Unread only</label
        >
        <button
          class="btn btn-ghost btn-xs"
          :disabled="!feed.unreadCount.value"
          @click="feed.markAllRead"
        >
          <CheckCheck :size="14" /> Mark all read
        </button>
      </div>
      <div
        class="max-h-[65vh] overflow-y-auto overscroll-contain"
        :aria-busy="loading"
      >
        <p
          v-if="loading && !feed.items.value.length"
          class="p-4 text-sm text-base-content/60"
        >
          Loading notifications…
        </p>
        <p
          v-else-if="!visible.length"
          class="p-4 text-sm text-base-content/60"
        >
          {{
            unreadOnly
              ? 'No unread notifications.'
              : 'No notifications right now.'
          }}
        </p>
        <div
          v-for="item in visible"
          :key="item.id"
          class="flex items-start gap-2 border-b border-base-300 p-3 last:border-0"
        >
          <span
            class="mt-2 size-1.5 shrink-0 rounded-full"
            :class="feed.isRead(item.id) ? 'bg-base-300' : 'bg-primary'"
            aria-hidden="true"
          ></span>
          <button
            class="min-w-0 flex-1 text-start hover:text-primary"
            @click="activate(item.id, item.destination)"
          >
            <span class="block text-sm font-semibold">{{ item.title }}</span>
            <span class="block break-words text-xs text-base-content/60">{{
              item.detail
            }}</span>
            <span v-if="item.date" class="block text-xs text-base-content/60">{{
              formatDateTime(item.date)
            }}</span>
            <span
              v-if="item.amountRial !== undefined"
              class="block text-xs font-medium"
              >{{ formatMoney(item.amountRial, currencyUnit) }}</span
            >
          </button>
          <button
            v-if="!feed.isRead(item.id)"
            class="btn btn-ghost btn-square btn-xs"
            :aria-label="`Mark ${item.title} as read`"
            @click="feed.markRead(item.id)"
          >
            <CheckCheck :size="14" />
          </button>
        </div>
      </div>
    </section>
  </div>
</template>
