<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { CalendarDays, ChevronLeft, ChevronRight, X } from 'lucide-vue-next'
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
} from '../utils/date'

const props = withDefaults(defineProps<{ modelValue: string | null; placeholder?: string }>(), { placeholder: 'Select Jalali date' })

const emit = defineEmits<{
  'update:modelValue': [value: string | null]
}>()

const picker = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const today = currentCanonicalDate()
const calendarMonth = ref<JalaliDate>(toJalaliDate(today))

const displayValue = computed(() => props.modelValue ? formatDate(props.modelValue) : '')
const monthLabel = computed(() => formatJalaliMonth(calendarMonth.value.year, calendarMonth.value.month))
const dayCells = computed(() => {
  const days = jalaliMonthLength(calendarMonth.value.year, calendarMonth.value.month)
  const offset = jalaliMonthStartOffset(calendarMonth.value.year, calendarMonth.value.month)
  return Array.from({ length: offset + days }, (_, index) => index < offset ? null : index - offset + 1)
})

function openPicker() {
  if (!isOpen.value && props.modelValue) calendarMonth.value = toJalaliDate(props.modelValue)
  isOpen.value = !isOpen.value
}

function shiftMonth(delta: number) {
  const nextMonth = calendarMonth.value.month + delta
  if (nextMonth < 1) {
    calendarMonth.value = { year: calendarMonth.value.year - 1, month: 12, day: 1 }
  } else if (nextMonth > 12) {
    calendarMonth.value = { year: calendarMonth.value.year + 1, month: 1, day: 1 }
  } else {
    calendarMonth.value = { ...calendarMonth.value, month: nextMonth, day: 1 }
  }
}

function selectDay(day: number) {
  const canonical = fromJalaliDate({ ...calendarMonth.value, day })
  if (!canonical) return
  emit('update:modelValue', canonical)
  isOpen.value = false
}

function selectToday() {
  calendarMonth.value = toJalaliDate(today)
  emit('update:modelValue', today)
  isOpen.value = false
}

function clearDate() {
  emit('update:modelValue', null)
  isOpen.value = false
}

function isSelected(day: number) {
  if (!props.modelValue) return false
  const selected = toJalaliDate(props.modelValue)
  return selected.year === calendarMonth.value.year && selected.month === calendarMonth.value.month && selected.day === day
}

function isToday(day: number) {
  const current = toJalaliDate(today)
  return current.year === calendarMonth.value.year && current.month === calendarMonth.value.month && current.day === day
}

function onDocumentClick(event: MouseEvent) {
  if (picker.value && !picker.value.contains(event.target as Node)) isOpen.value = false
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') isOpen.value = false
}

watch(() => props.modelValue, (value) => {
  if (value && isOpen.value) calendarMonth.value = toJalaliDate(value)
})

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div ref="picker" class="jalali-picker">
    <div class="jalali-picker-control">
      <input
        :value="displayValue"
        class="jalali-picker-input"
        type="text"
        readonly
        :placeholder="placeholder"
        aria-label="Promised date"
        :aria-expanded="isOpen"
        aria-haspopup="dialog"
        @click="openPicker"
        @keydown.enter.space.prevent="openPicker"
        @keydown.down.prevent="openPicker"
      />
      <button class="jalali-picker-trigger" type="button" aria-label="Open Jalali calendar" :aria-expanded="isOpen" @click="openPicker">
        <CalendarDays :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
    </div>

    <div v-if="isOpen" class="jalali-picker-popover" role="dialog" aria-label="Jalali calendar">
      <div class="jalali-picker-heading">
        <button class="jalali-picker-nav" type="button" aria-label="Previous Jalali month" @click="shiftMonth(-1)"><ChevronLeft :size="16" :stroke-width="1.8" aria-hidden="true" /></button>
        <strong>{{ monthLabel }}</strong>
        <button class="jalali-picker-nav" type="button" aria-label="Next Jalali month" @click="shiftMonth(1)"><ChevronRight :size="16" :stroke-width="1.8" aria-hidden="true" /></button>
      </div>
      <div class="jalali-picker-weekdays" aria-hidden="true">
        <span v-for="weekday in jalaliWeekdays" :key="weekday">{{ weekday }}</span>
      </div>
      <div class="jalali-picker-days" role="grid" :aria-label="monthLabel">
        <span v-for="(day, index) in dayCells" :key="`${monthLabel}-${index}`" class="jalali-picker-cell">
          <button v-if="day" class="jalali-picker-day" :class="{ 'is-selected': isSelected(day), 'is-today': isToday(day) }" type="button" role="gridcell" :aria-label="`${day} ${monthLabel}`" :aria-selected="isSelected(day)" @click="selectDay(day)">{{ day }}</button>
        </span>
      </div>
      <div class="jalali-picker-footer">
        <button class="text-button" type="button" @click="selectToday">Today</button>
        <button v-if="modelValue" class="jalali-picker-clear" type="button" aria-label="Clear promised date" @click="clearDate"><X :size="14" :stroke-width="1.8" aria-hidden="true" />Clear</button>
      </div>
    </div>
  </div>
</template>
