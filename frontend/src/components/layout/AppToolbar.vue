<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  BarChart3,
  ChevronDown,
  FilePlus2,
  PanelLeftClose,
  PanelLeftOpen,
  Printer,
  ShoppingCart,
  Store,
} from 'lucide-vue-next';
import type { CurrencyUnit } from '../../utils/currency';
import IconButton from '../ui/IconButton.vue';
import SearchField from '../ui/SearchField.vue';
import SelectField from '../ui/SelectField.vue';
import { messages, type Locale } from '../../i18n/messages';

interface GlobalSearchResult {
  id: string;
  view: string;
  title: string;
  subtitle: string;
  detail: string;
}

const searchField = ref<{ focus: () => void } | null>(null);
const { t } = useI18n();
const currencyOptions = computed(() => [
  { label: t('toolbar.toman'), value: 'Toman' as const },
  { label: t('toolbar.rial'), value: 'Rial' as const },
]);
const languageOptions = computed(() => [
  { label: t('language.english'), value: 'en' as const },
  { label: t('language.persian'), value: 'fa' as const },
]);

function translateView(view: string) {
  if (Object.hasOwn(messages.en.navigation.views, view)) {
    return t(`navigation.views.${view as keyof typeof messages.en.navigation.views}`);
  }
  return view;
}

function focusSearch(event: KeyboardEvent) {
  if (event.ctrlKey && event.key.toLowerCase() === 'f') {
    event.preventDefault();
    searchField.value?.focus();
  }
}

onMounted(() => window.addEventListener('keydown', focusSearch));
onBeforeUnmount(() => window.removeEventListener('keydown', focusSearch));

defineProps<{
  collapsed: boolean;
  searchQuery: string;
  searchResults: GlobalSearchResult[];
  searchLoading: boolean;
  currencyUnit: CurrencyUnit;
  locale: Locale;
}>();

defineEmits<{
  'toggle-sidebar': [];
  'update:search-query': [value: string];
  'select-search-result': [result: GlobalSearchResult];
  'update:currency-unit': [value: CurrencyUnit];
  'update:locale': [value: Locale];
  'new-order': [];
  navigate: [view: string];
}>();
</script>

<template>
  <header class="navbar min-h-16 gap-3 border-b border-base-300 bg-base-100 px-4 py-3 lg:px-6">
    <IconButton
      data-drawer-toggle="true"
      class="btn-square"
      :label="collapsed ? t('toolbar.expandSidebar') : t('toolbar.collapseSidebar')"
      @click.stop="$emit('toggle-sidebar')"
    >
      <PanelLeftOpen v-if="collapsed" :size="18" :stroke-width="1.8" aria-hidden="true" />
      <PanelLeftClose v-else :size="18" :stroke-width="1.8" aria-hidden="true" />
    </IconButton>
    <div class="relative min-w-0 flex-1">
    <SearchField
      ref="searchField"
      class="w-full"
      :model-value="searchQuery"
      :placeholder="t('toolbar.search')"
      shortcut="Ctrl F"
      :aria-label="t('toolbar.globalSearch')"
      @update:model-value="$emit('update:search-query', $event)"
      @keydown.esc="$emit('update:search-query', '')"
      @keydown.enter.prevent="searchResults[0] && $emit('select-search-result', searchResults[0])"
    />
    <div
      v-if="searchQuery.trim()"
      class="absolute inset-x-0 top-[calc(100%+0.45rem)] z-50 overflow-hidden rounded-box border border-base-300 bg-base-100 shadow-2xl"
    >
      <div v-if="searchLoading && !searchResults.length" class="px-4 py-3 text-xs text-base-content/60">
        {{ t('toolbar.searching') }}
      </div>
      <div v-else-if="!searchResults.length" class="px-4 py-3 text-xs text-base-content/60">
        {{ t('toolbar.noResults') }}
      </div>
      <div v-else class="max-h-[min(28rem,calc(100vh-6rem))] overflow-y-auto p-1">
        <button
          v-for="result in searchResults"
          :key="`${result.view}-${result.id}`"
          class="flex w-full min-w-0 items-start gap-3 rounded-box px-3 py-2.5 text-start transition-colors hover:bg-base-200 focus:bg-base-200 focus:outline-none"
          type="button"
          @mousedown.prevent
          @click="$emit('select-search-result', result)"
        >
          <span class="mt-0.5 grid size-7 shrink-0 place-items-center rounded bg-primary/10 text-[10px] font-bold text-primary">
            {{ result.view.slice(0, 1) }}
          </span>
          <span class="min-w-0 flex-1">
            <span class="flex min-w-0 items-center justify-between gap-3">
              <strong class="truncate text-sm">{{ result.title }}</strong>
              <span class="shrink-0 text-[10px] uppercase tracking-wide text-base-content/45">{{ translateView(result.view) }}</span>
            </span>
            <span class="mt-0.5 block truncate text-xs text-base-content/60">{{ result.subtitle }}</span>
            <span v-if="result.detail" class="mt-0.5 block truncate text-[11px] text-base-content/45">{{ result.detail }}</span>
          </span>
        </button>
      </div>
    </div>
    </div>
    <button class="btn btn-ghost hidden" type="button" :aria-label='$t("Current shop: Central shop")'>
      <Store class="text-primary" :size="16" :stroke-width="1.8" aria-hidden="true" />
      <span><strong>{{ $t("Central shop") }}</strong> {{ $t("· Tehran") }}</span>
      <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" />
    </button>
    <SelectField
      class="w-24"
      :model-value="currencyUnit"
      :options="currencyOptions"
      :aria-label="t('toolbar.currency')"
      @update:model-value="$emit('update:currency-unit', $event as CurrencyUnit)"
    />
    <SelectField
      class="w-20 shrink-0"
      :model-value="locale"
      :options="languageOptions"
      :aria-label="t('language.label')"
      @update:model-value="$emit('update:locale', $event as Locale)"
    />
    <slot name="notifications" />
    <nav class="hidden items-center gap-1 lg:flex" :aria-label="t('toolbar.quickActions')">
      <button
        class="btn btn-ghost btn-square"
        type="button"
        :aria-label="t('navigation.views.Production')"
        :title="t('navigation.views.Production')"
        @click="$emit('navigate', 'Production')"
      >
        <Printer :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <button
        class="btn btn-ghost btn-square"
        type="button"
        :aria-label="t('navigation.views.Purchases')"
        :title="t('navigation.views.Purchases')"
        @click="$emit('navigate', 'Purchases')"
      >
        <ShoppingCart :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <button
        class="btn btn-ghost btn-square"
        type="button"
        :aria-label="t('navigation.views.Reports')"
        :title="t('navigation.views.Reports')"
        @click="$emit('navigate', 'Reports')"
      >
        <BarChart3 :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <button
        class="btn btn-primary btn-square"
        type="button"
        :aria-label="t('toolbar.newOrder')"
        :title="t('toolbar.newOrder')"
        @click="$emit('new-order')"
      >
        <FilePlus2 :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
    </nav>
    <div class="avatar placeholder hidden">
      <div class="w-8 rounded-full bg-primary text-primary-content" aria-hidden="true">
        <span>{{ $t("RR") }}</span>
      </div>
    </div>
    <div class="hidden text-xs">
      <strong>{{ $t("Reza Rasoulzadeh") }}</strong><span class="block text-base-content/60">{{ $t("Owner") }}</span>
    </div>
  </header>
</template>
