<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Plus, RotateCcw, Save, Trash2, UserRound } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import AppPanel from '../../components/layout/AppPanel.vue'
import DataTable from '../../components/ui/DataTable.vue'
import DataTableCell from '../../components/ui/DataTableCell.vue'
import DataTableRow from '../../components/ui/DataTableRow.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import FormField from '../../components/ui/FormField.vue'
import FormGrid from '../../components/ui/FormGrid.vue'
import InspectorShell from '../../components/layout/InspectorShell.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import MasterDetail from '../../components/layout/MasterDetail.vue'
import SearchField from '../../components/ui/SearchField.vue'
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue'
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue'
import { useWorkspaceActions, reportError } from '../../composables/useWorkspaceActions'
import { ownersApi, type OwnerRecord } from '../../api/owners'
import { formatMoney, type CurrencyUnit } from '../../utils/currency'
import { confirmAction } from '../../ui/feedback'

const props = defineProps<{ currencyUnit: CurrencyUnit }>()
const emit = defineEmits<{ notify: [string] }>()
const { busy, pageLoading, runAction, runLoad } = useWorkspaceActions()

const owners = ref<OwnerRecord[]>([])
const selectedId = ref<string | null>(null)
const query = ref('')
const editing = ref(false)
const saving = ref(false)
const ownerForm = ref({ name: '', phone: '', email: '', notes: '', ownershipBps: 0, profitSharingBps: 0 })
const shareForm = ref({ ownershipBps: 0, profitSharingBps: 0 })
const selected = computed(() => owners.value.find((owner) => owner.id === selectedId.value) ?? null)
const visible = computed(() => {
  const search = query.value.trim().toLowerCase()
  return owners.value.filter((owner) => !search || [owner.name, owner.phone, owner.email].join(' ').toLowerCase().includes(search))
})

function money(value: number) {
  return formatMoney(value, props.currencyUnit)
}

function syncShares() {
  if (selected.value) {
    shareForm.value = {
      ownershipBps: selected.value.ownershipBps,
      profitSharingBps: selected.value.profitSharingBps,
    }
  }
}

function select(owner: OwnerRecord) {
  selectedId.value = owner.id
  editing.value = false
  syncShares()
}

async function load() {
  return runLoad(async () => {
    try {
      owners.value = await ownersApi.list(false)
      if (!selectedId.value && owners.value[0]) select(owners.value[0])
    } catch (error) {
      reportError(error)
    }
  })
}

onMounted(load)

function startOwner() {
  ownerForm.value = { name: '', phone: '', email: '', notes: '', ownershipBps: 0, profitSharingBps: 0 }
  editing.value = true
  selectedId.value = null
}

async function saveOwner() {
  return runAction(async () => {
    saving.value = true
    try {
      const value = await ownersApi.create(ownerForm.value)
      owners.value = [value, ...owners.value]
      selectedId.value = value.id
      editing.value = false
      syncShares()
      emit('notify', 'Owner saved.')
    } catch (error) {
      reportError(error)
    } finally {
      saving.value = false
    }
  })
}

async function saveShares() {
  return runAction(async () => {
    if (!selected.value) return
    try {
      const value = await ownersApi.updateShares(selected.value.id, shareForm.value)
      owners.value = owners.value.map((owner) => (owner.id === value.id ? value : owner))
      emit('notify', 'Owner shares updated.')
    } catch (error) {
      reportError(error)
    }
  })
}

async function toggleActive() {
  return runAction(async () => {
    if (!selected.value) return
    try {
      if (selected.value.active) await ownersApi.archive(selected.value.id)
      else await ownersApi.reactivate(selected.value.id)
      await load()
      emit('notify', selected.value?.active ? 'Owner reactivated.' : 'Owner archived.')
    } catch (error) {
      reportError(error)
    }
  })
}

async function removeOwner() {
  return runAction(async () => {
    if (
      !selected.value ||
      !(await confirmAction({
        title: 'Delete owner',
        message: 'Delete this owner permanently when safe?',
        confirmLabel: 'Delete owner',
        danger: true,
      }))
    )
      return
    try {
      await ownersApi.delete(selected.value.id)
      owners.value = owners.value.filter((owner) => owner.id !== selected.value!.id)
      selectedId.value = owners.value[0]?.id ?? null
      syncShares()
      emit('notify', 'Owner deleted.')
    } catch (error) {
      reportError(error)
    }
  })
}
</script>

<template>
  <div class="min-w-0 space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        :show-breadcrumb="true"
        eyebrow="Finance / ownership"
        title="Owners"
        description="Manage ownership shares, profit shares, and owner profile information."
      >
        <button class="btn btn-primary" type="button" @click="startOwner"><Plus :size="16" /> New owner</button>
      </WorkspaceHeader>
      <SearchFilterBar>
        <template #search><SearchField v-model="query" label="Search owners" placeholder="Name, phone, or email" /></template>
        <template #count><span>{{ visible.length }} of {{ owners.length }} owners</span></template>
      </SearchFilterBar>
    </WorkspaceStickyStack>

    <LoadingState v-if="pageLoading" label="Loading owners…" />
    <MasterDetail v-else wide>
      <AppPanel title="Owner register" subtitle="Ownership and profit-sharing percentages are stored as basis points." :flush="true">
        <EmptyState v-if="!visible.length" title="No owners in this view" :description="owners.length ? 'Adjust the search filter.' : 'Create an owner to begin tracking ownership.'">
          <template #icon><UserRound :size="22" aria-hidden="true" /></template>
          <template #action>
            <button v-if="owners.length" class="btn btn-primary btn-sm" type="button" @click="query = ''">Clear filters</button>
            <button v-else class="btn btn-primary btn-sm gap-2" type="button" @click="startOwner"><Plus :size="15" aria-hidden="true" /> Create owner</button>
          </template>
        </EmptyState>
        <DataTable v-else label="Owner register">
          <thead><tr><th scope="col" class="w-[45%]">Owner</th><th scope="col" class="w-[25%]">Shares</th><th scope="col" class="w-[30%] text-end">Current balance</th></tr></thead>
          <tbody>
            <DataTableRow v-for="owner in visible" :key="owner.id" interactive :selected="selectedId === owner.id" @activate="select(owner)">
              <DataTableCell><div class="flex min-w-0 items-center gap-2"><div class="min-w-0"><strong class="block truncate text-sm">{{ owner.name }}</strong><span class="mt-1 block truncate text-xs text-base-content/55">{{ owner.email || owner.phone || 'No contact details' }}</span></div><StatusBadge class="shrink-0" :label="owner.active ? 'Active' : 'Archived'" :tone="owner.active ? 'green' : 'slate'" /></div></DataTableCell>
              <DataTableCell><div class="grid min-w-24 grid-cols-2 gap-2 text-xs"><span><small class="block text-base-content/50">Own</small><strong class="block tabular-nums">{{ owner.ownershipBps / 100 }}%</strong></span><span><small class="block text-base-content/50">P/L</small><strong class="block tabular-nums">{{ owner.profitSharingBps / 100 }}%</strong></span></div></DataTableCell>
              <DataTableCell numeric><strong>{{ money(owner.currentBalanceRial) }}</strong></DataTableCell>
            </DataTableRow>
          </tbody>
        </DataTable>
      </AppPanel>

      <InspectorShell v-if="editing" title="New owner" subtitle="Add a person to the ownership register.">
        <form class="min-w-0 space-y-4" @submit.prevent="saveOwner">
          <FormField label="Name"><AppInput v-model="ownerForm.name" class="input w-full min-w-0" required placeholder="Owner name" /></FormField>
          <FormGrid>
            <FormField label="Ownership % (bps)"><AppInput v-model.number="ownerForm.ownershipBps" class="input w-full min-w-0" type="number" min="0" max="10000" required /></FormField>
            <FormField label="Profit share % (bps)"><AppInput v-model.number="ownerForm.profitSharingBps" class="input w-full min-w-0" type="number" min="0" max="10000" required /></FormField>
            <FormField label="Phone"><AppInput v-model="ownerForm.phone" class="input w-full min-w-0" /></FormField>
            <FormField label="Email"><AppInput v-model="ownerForm.email" class="input w-full min-w-0" type="email" /></FormField>
          </FormGrid>
          <FormField label="Notes"><AppInput v-model="ownerForm.notes" class="input w-full min-w-0" /></FormField>
          <div class="flex justify-end gap-2"><button class="btn btn-ghost" type="button" @click="editing = false">Cancel</button><button class="btn btn-primary" type="submit" :disabled="busy || saving"><Save :size="15" /> Save owner</button></div>
        </form>
      </InspectorShell>

      <InspectorShell v-else-if="selected" :title="selected.name" subtitle="Profile, shares, and derived owner balances.">
        <template #header><StatusBadge :label="selected.active ? 'Active' : 'Archived'" :tone="selected.active ? 'green' : 'slate'" /></template>
        <div class="min-w-0 space-y-5">
          <section class="flex min-w-0 items-start gap-3 rounded-box border border-primary/20 bg-primary/5 p-4">
            <div class="grid size-10 shrink-0 place-items-center rounded-box bg-primary/15 text-primary"><UserRound :size="19" /></div>
            <div class="min-w-0"><p class="text-xs font-medium uppercase tracking-wide text-primary/80">Owner profile</p><p class="mt-1 text-lg font-bold">{{ selected.name }}</p><p class="mt-1 text-xs text-base-content/60">{{ selected.email || selected.phone || 'No contact details' }}</p></div>
          </section>
          <section class="grid min-w-0 grid-cols-2 gap-2">
            <div class="rounded-box border border-base-300 bg-base-200/30 p-3"><span class="block text-xs text-base-content/55">Ownership</span><strong class="mt-1 block text-lg tabular-nums">{{ selected.ownershipBps / 100 }}%</strong></div>
            <div class="rounded-box border border-base-300 bg-base-200/30 p-3"><span class="block text-xs text-base-content/55">Profit share</span><strong class="mt-1 block text-lg tabular-nums">{{ selected.profitSharingBps / 100 }}%</strong></div>
          </section>
          <form class="space-y-3" @submit.prevent="saveShares">
            <h3 class="text-sm font-semibold">Future share settings</h3>
            <FormGrid>
              <FormField label="Ownership bps"><AppInput v-model.number="shareForm.ownershipBps" class="input w-full min-w-0" type="number" min="0" max="10000" /></FormField>
              <FormField label="Profit share bps"><AppInput v-model.number="shareForm.profitSharingBps" class="input w-full min-w-0" type="number" min="0" max="10000" /></FormField>
            </FormGrid>
            <button class="btn btn-primary btn-sm" type="submit" :disabled="busy"><Save :size="14" /> Save shares</button>
          </form>
          <dl class="grid min-w-0 gap-2 rounded-box border border-base-300 bg-base-200/20 p-3 text-sm">
            <div class="flex justify-between gap-3 border-b border-base-300 py-2"><dt class="text-xs text-base-content/60">Capital contributed</dt><dd class="tabular-nums">{{ money(selected.capitalContributedRial) }}</dd></div>
            <div class="flex justify-between gap-3 border-b border-base-300 py-2"><dt class="text-xs text-base-content/60">Drawings</dt><dd class="tabular-nums">{{ money(selected.drawingsRial) }}</dd></div>
            <div class="flex justify-between gap-3 border-b border-base-300 py-2"><dt class="text-xs text-base-content/60">Current balance</dt><dd class="font-semibold tabular-nums text-primary">{{ money(selected.currentBalanceRial) }}</dd></div>
            <div class="flex justify-between gap-3 py-2"><dt class="text-xs text-base-content/60">Allocated P/L</dt><dd class="tabular-nums">{{ money(selected.allocatedProfitLossRial) }}</dd></div>
          </dl>
        </div>
        <div class="flex flex-wrap justify-end gap-2 border-t border-base-300 pt-4"><button class="btn btn-ghost btn-sm" type="button" :disabled="busy" @click="toggleActive"><RotateCcw :size="14" /> {{ selected.active ? 'Archive' : 'Reactivate' }}</button><button class="btn btn-outline btn-error btn-sm" type="button" :disabled="busy" @click="removeOwner"><Trash2 :size="14" /> Delete</button></div>
      </InspectorShell>

      <InspectorShell v-else title="Owner inspector" subtitle="Select an owner from the register."><EmptyState title="Select an owner" description="Choose a row to inspect shares and balances."><template #icon><UserRound :size="21" aria-hidden="true" /></template></EmptyState></InspectorShell>
    </MasterDetail>
  </div>
</template>
