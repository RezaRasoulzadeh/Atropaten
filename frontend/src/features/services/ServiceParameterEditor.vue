<script setup lang="ts">
import FormGrid from '../../components/ui/FormGrid.vue'
import FormField from '../../components/ui/FormField.vue'
import AppInput from '../../components/ui/AppInput.vue'
import SelectField from '../../components/ui/SelectField.vue'
import {ChevronUp,ChevronDown,Trash2,Plus} from 'lucide-vue-next'
import type {ParameterForm} from './types'
import type {MaterialRecord} from '../../api/materials'
defineProps<{parameter:ParameterForm;index:number;count:number;materials:MaterialRecord[]}>()
const emit=defineEmits<{move:[direction:-1|1];remove:[];normalize:[];addOption:[];removeOption:[index:number]}>()
</script>
<template><article class="min-w-0 space-y-3 border-t border-base-300 py-4">
              <header class="flex min-w-0 flex-wrap items-start justify-between gap-3">
                <span>{{ String(index + 1).padStart(2, '0') }}</span
                ><strong>{{ parameter.label || 'Untitled parameter' }}</strong>
                <div class="flex flex-wrap items-center gap-2">
                  <button
                    class="btn btn-ghost"
                    type="button"
                    :disabled="index === 0"
                    :aria-label="`Move ${parameter.label || 'parameter'} up`"
                    @click="emit('move',-1)"
                  >
                    <ChevronUp :size="14" :stroke-width="1.8" aria-hidden="true" /></button
                  ><button
                    class="btn btn-ghost"
                    type="button"
                    :disabled="index === count - 1"
                    :aria-label="`Move ${parameter.label || 'parameter'} down`"
                    @click="emit('move',1)"
                  >
                    <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" /></button
                  ><button
                    class="btn btn-ghost"
                    type="button"
                    :aria-label="`Remove ${parameter.label || 'parameter'}`"
                    @click="emit('remove')"
                  >
                    <Trash2 :size="14" :stroke-width="1.8" aria-hidden="true" />
                  </button>
                </div>
              </header>
              <FormGrid
                ><FormField class="gap-1"
                  ><span>Key</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="parameter.key"
                    type="text"
                    placeholder="paper_size"
                    autocomplete="off" /></FormField
                ><FormField class="gap-1"
                  ><span>Label</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="parameter.label"
                    type="text"
                    placeholder="Paper size"
                    autocomplete="off" /></FormField
              ></FormGrid>
              <FormGrid
                ><SelectField
                  v-model="parameter.type"
                  label="Type"
                  :options="[
                    { label: 'Integer', value: 'integer' },
                    { label: 'Decimal', value: 'decimal' },
                    { label: 'Boolean', value: 'boolean' },
                    { label: 'Choice', value: 'choice' },
                    { label: 'Material reference', value: 'material-reference' },
                  ]"
                  @update:model-value="emit('normalize')"
                /><FormField class="gap-1"
                  ><span>Required</span
                  ><span
                    ><input class="checkbox" v-model="parameter.required" type="checkbox" />Required
                    input</span
                  ></FormField
                ></FormGrid
              >
              <FormGrid v-if="parameter.type === 'integer' || parameter.type === 'decimal'"
                ><FormField class="gap-1"
                  ><span>Minimum</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="parameter.minValue"
                    type="text"
                    inputmode="decimal"
                    placeholder="Optional" /></FormField
                ><FormField class="gap-1"
                  ><span>Maximum</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="parameter.maxValue"
                    type="text"
                    inputmode="decimal"
                    placeholder="Optional" /></FormField
              ></FormGrid>
              <FormField
                class="gap-1"
                v-if="parameter.type === 'integer' || parameter.type === 'decimal'"
                ><span>Default value</span
                ><AppInput
                  class="input w-full min-w-0"
                  v-model="parameter.defaultValue"
                  type="text"
                  inputmode="decimal"
                  placeholder="Optional"
              /></FormField>
              <SelectField
                v-else-if="parameter.type === 'boolean'"
                v-model="parameter.defaultValue"
                label="Default value"
                :options="[
                  { label: 'Not set', value: '' },
                  { label: 'True', value: 'true' },
                  { label: 'False', value: 'false' },
                ]"
              />
              <SelectField
                v-else-if="parameter.type === 'choice'"
                v-model="parameter.defaultValue"
                label="Default choice"
                :options="[
                  { label: 'Not set', value: '' },
                  ...parameter.options.map((option) => ({ label: option, value: option })),
                ]"
              />
              <div v-else class="min-w-0 space-y-3">
                <SelectField
                  v-model="parameter.defaultValue"
                  label="Default material"
                  :options="[
                    { label: 'No default material', value: '' },
                    ...materials.map((material) => ({
                      label: `${material.name}${material.sku ? ` · ${material.sku}` : ''}`,
                      value: material.id,
                    })),
                  ]"
                /><small class="block text-xs leading-5 text-base-content/60"
                  >Operators may select an active material later; consumption is not configured
                  here.</small
                >
              </div>
              <div v-if="parameter.type === 'choice'" class="min-w-0 space-y-3">
                <div class="min-w-0 space-y-3">
                  <span>Choice options</span
                  ><button class="btn btn-ghost" type="button" @click="emit('addOption')">
                    <Plus :size="14" :stroke-width="1.8" aria-hidden="true" />Add option
                  </button>
                </div>
                <div
                  v-for="(option, optionIndex) in parameter.options"
                  :key="`${parameter.id}-${optionIndex}`"
                  class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
                >
                  <AppInput
                    class="input w-full min-w-0"
                    v-model="parameter.options[optionIndex]"
                    type="text"
                    :aria-label="`Choice option ${optionIndex + 1}`"
                    placeholder="A4"
                    @input="
                      parameter.defaultValue =
                        parameter.defaultValue === option
                          ? parameter.options[optionIndex]
                          : parameter.defaultValue
                    "
                  /><button
                    class="btn btn-ghost"
                    type="button"
                    :aria-label="`Remove choice option ${optionIndex + 1}`"
                    @click="emit('removeOption',optionIndex)"
                  >
                    <Trash2 :size="13" :stroke-width="1.8" aria-hidden="true" />
                  </button>
                </div>
                <p v-if="!parameter.options.length">Add at least one option before saving.</p>
              </div>
              <FormField class="gap-1"
                ><span>Unit / suffix <em>optional</em></span
                ><AppInput
                  class="input w-full min-w-0"
                  v-model="parameter.unit"
                  type="text"
                  placeholder="sheets, hours, or mm"
              /></FormField>
            </article></template>