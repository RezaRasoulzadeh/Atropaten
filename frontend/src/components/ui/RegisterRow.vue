<script setup lang="ts">
withDefaults(defineProps<{
  selected?: boolean;
  interactive?: boolean;
}>(), { interactive: true });

const emit = defineEmits<{
  activate: [];
}>();
</script>

<template>
  <article
    class="flex min-w-0 items-center gap-3 px-4 py-3 text-start transition-colors hover:bg-base-200 focus-visible:bg-base-200 focus-visible:outline focus-visible:outline-1 focus-visible:outline-base-content focus-visible:outline-offset-[-1px]"
    :class="selected ? 'bg-base-300' : ''"
     :role="interactive === false ? undefined : 'button'"
    :tabindex="interactive === false ? undefined : 0"
    @click="interactive !== false && emit('activate')"
    @keydown.enter.self.prevent="interactive !== false && emit('activate')"
    @keydown.space.self.prevent="interactive !== false && emit('activate')"
  >
    <div v-if="$slots.icon"
      class="grid size-9 shrink-0 place-items-center rounded-box bg-base-200 text-base-content/65"
    >
      <slot name="icon" />
    </div>
    <div class="min-w-0 flex-1">
      <slot name="identity" />
      <slot name="meta" />
    </div>
    <div class="flex shrink-0 items-center gap-2">
      <slot name="status" />
      <slot name="actions" />
    </div>
  </article>
</template>
