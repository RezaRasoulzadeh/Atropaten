<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { CalendarDays, ChevronLeft, ChevronRight, X } from 'lucide-vue-next';
import {
  currentCanonicalDate,
  formatDate,
  formatJalaliMonth,
  fromJalaliDate,
  jalaliMonthLength,
  jalaliMonthStartOffset,
  jalaliWeekdays,
  toJalaliDate,
  type JalaliDate,
} from '../../utils/date';

const props = withDefaults(defineProps<{ modelValue: string | null; placeholder?: string; disabled?: boolean }>(), {
  placeholder: 'Select Jalali date',
  disabled: false,
});

const emit = defineEmits<{
  'update:modelValue': [value: string | null];
}>();

const picker = ref<HTMLElement | null>(null);
const popover = ref<HTMLElement | null>(null);
const isOpen = ref(false);
const popoverStyle = ref({ top: '0px', left: '0px' });
const today = currentCanonicalDate();
const calendarMonth = ref<JalaliDate>(toJalaliDate(today));

const displayValue = computed(() => (props.modelValue ? formatDate(props.modelValue) : ''));
const monthLabel = computed(() =>
  formatJalaliMonth(calendarMonth.value.year, calendarMonth.value.month),
);
const dayCells = computed(() => {
  const days = jalaliMonthLength(calendarMonth.value.year, calendarMonth.value.month);
  const offset = jalaliMonthStartOffset(calendarMonth.value.year, calendarMonth.value.month);
  return Array.from({ length: offset + days }, (_, index) =>
    index < offset ? null : index - offset + 1,
  );
});

function openPicker() {
  if (props.disabled) return
  if (!isOpen.value && props.modelValue) calendarMonth.value = toJalaliDate(props.modelValue);
  isOpen.value = !isOpen.value;
  if (isOpen.value) void nextTick(positionPopover);
}

function positionPopover() {
  if (!isOpen.value || !picker.value || !popover.value) return;
  const anchor = picker.value.getBoundingClientRect();
  const menu = popover.value.getBoundingClientRect();
  const edge = 8;
  const below = anchor.bottom + edge + menu.height <= window.innerHeight - edge;
  const top = below ? anchor.bottom + edge : Math.max(edge, anchor.top - menu.height - edge);
  const left = Math.min(
    Math.max(edge, anchor.left),
    Math.max(edge, window.innerWidth - menu.width - edge),
  );
  popoverStyle.value = { top: `${Math.round(top)}px`, left: `${Math.round(left)}px` };
}

function shiftMonth(delta: number) {
  const nextMonth = calendarMonth.value.month + delta;
  if (nextMonth < 1) {
    calendarMonth.value = { year: calendarMonth.value.year - 1, month: 12, day: 1 };
  } else if (nextMonth > 12) {
    calendarMonth.value = { year: calendarMonth.value.year + 1, month: 1, day: 1 };
  } else {
    calendarMonth.value = { ...calendarMonth.value, month: nextMonth, day: 1 };
  }
}

function selectDay(day: number) {
  const canonical = fromJalaliDate({ ...calendarMonth.value, day });
  if (!canonical) return;
  emit('update:modelValue', canonical);
  isOpen.value = false;
}

function selectToday() {
  calendarMonth.value = toJalaliDate(today);
  emit('update:modelValue', today);
  isOpen.value = false;
}

function clearDate() {
  emit('update:modelValue', null);
  isOpen.value = false;
}

function isSelected(day: number) {
  if (!props.modelValue) return false;
  const selected = toJalaliDate(props.modelValue);
  return (
    selected.year === calendarMonth.value.year &&
    selected.month === calendarMonth.value.month &&
    selected.day === day
  );
}

function isToday(day: number) {
  const current = toJalaliDate(today);
  return (
    current.year === calendarMonth.value.year &&
    current.month === calendarMonth.value.month &&
    current.day === day
  );
}

function onDocumentClick(event: MouseEvent) {
  const target = event.target as Node;
  if (
    picker.value &&
    !picker.value.contains(target) &&
    !popover.value?.contains(target)
  ) {
    isOpen.value = false;
  }
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') isOpen.value = false;
}

watch(
  () => props.modelValue,
  (value) => {
    if (value && isOpen.value) calendarMonth.value = toJalaliDate(value);
  },
);

onMounted(() => {
  document.addEventListener('click', onDocumentClick);
  document.addEventListener('keydown', onKeydown);
  window.addEventListener('resize', positionPopover);
  window.addEventListener('scroll', positionPopover, true);
});

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick);
  document.removeEventListener('keydown', onKeydown);
  window.removeEventListener('resize', positionPopover);
  window.removeEventListener('scroll', positionPopover, true);
});
</script>

<template>
  <div ref="picker" class="relative w-full">
    <div class="relative">
      <input
        class="input input-bordered w-full min-w-0 pe-11 leading-5"
        :value="displayValue"
        type="text"
        readonly
        :disabled="disabled"
        :placeholder="placeholder"
        aria-label="Promised date"
        :aria-expanded="isOpen"
        aria-haspopup="dialog"
        @click="openPicker"
        @keydown.enter.space.prevent="openPicker"
        @keydown.down.prevent="openPicker"
      />
      <button
        class="btn btn-ghost btn-square btn-sm absolute inset-e-1 top-1/2 -translate-y-1/2"
        type="button"
        :disabled="disabled"
        aria-label="Open Jalali calendar"
        :aria-expanded="isOpen"
        @click="openPicker"
      >
        <CalendarDays :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
    </div>

    <Teleport to="body">
      <div
        v-if="isOpen"
        ref="popover"
        class="fixed z-50 min-w-72 rounded-box border border-base-300 bg-base-100 p-3 shadow-xl"
        :style="popoverStyle"
        role="dialog"
        aria-label="Jalali calendar"
      >
      <div class="flex items-center justify-between gap-2">
        <button
          class="btn btn-ghost btn-square btn-sm"
          type="button"
          aria-label="Previous Jalali month"
          @click="shiftMonth(-1)"
        >
          <ChevronLeft :size="16" :stroke-width="1.8" aria-hidden="true" />
        </button>
        <strong class="text-sm">{{ monthLabel }}</strong>
        <button
          class="btn btn-ghost btn-square btn-sm"
          type="button"
          aria-label="Next Jalali month"
          @click="shiftMonth(1)"
        >
          <ChevronRight :size="16" :stroke-width="1.8" aria-hidden="true" />
        </button>
      </div>
      <div
        class="mt-2 grid grid-cols-7 gap-1 text-center text-xs leading-4 text-base-content/50"
        aria-hidden="true"
      >
        <span v-for="weekday in jalaliWeekdays" :key="weekday">{{ weekday }}</span>
      </div>
      <div class="mt-1 grid grid-cols-7 gap-1" role="grid" :aria-label="monthLabel">
        <span
          v-for="(day, index) in dayCells"
          :key="`${monthLabel}-${index}`"
          class="grid place-items-center"
        >
          <button
            v-if="day"
            class="btn btn-ghost btn-square btn-sm h-8 min-h-8 w-8 p-0 text-xs leading-4"
            :class="{
              'btn-primary': isSelected(day),
              'ring-1 ring-primary': isToday(day) && !isSelected(day),
            }"
            type="button"
            role="gridcell"
            :aria-label="`${day} ${monthLabel}`"
            :aria-selected="isSelected(day)"
            @click="selectDay(day)"
          >
            {{ day }}
          </button>
        </span>
      </div>
      <div class="mt-3 flex items-center justify-between border-t border-base-300 pt-2">
        <button class="btn btn-ghost btn-sm" type="button" @click="selectToday">Today</button>
        <button
          v-if="modelValue"
          class="btn btn-ghost btn-sm gap-1"
          type="button"
          aria-label="Clear promised date"
          @click="clearDate"
        >
          <X :size="14" :stroke-width="1.8" aria-hidden="true" />Clear
        </button>
      </div>
      </div>
    </Teleport>
  </div>
</template>
