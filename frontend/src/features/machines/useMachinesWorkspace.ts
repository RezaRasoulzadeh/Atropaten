import { computed, onMounted, ref, watch } from 'vue'
import { machinesApi, type MachinePayload, type MachineRatePayload, type MachineRecord } from '../../api/machines'
import { useWorkspaceActions } from '../../composables/useWorkspaceActions'
import { confirmAction, useToast } from '../../ui/feedback'
import { formatMoneyInput, parseMoneyInput, type CurrencyUnit } from '../../utils/currency'
import { formatDateTime } from '../../utils/date'

export type MachineFilter = 'Active' | 'Archived' | 'All'
export type MachineEditorMode = 'create' | 'edit' | null
export type MachineRateForm = Omit<MachineRatePayload, 'rateRial' | 'setupCostRial'> & {
  rate: string
  setupCost: string
}
export type MachineForm = Omit<MachinePayload, 'rateRial' | 'setupCostRial' | 'rates'> & {
  rate: string
  setupCost: string
  rates: MachineRateForm[]
}

type MachinesProps = { currencyUnit: CurrencyUnit }
type MachinesEmit = (event: 'notify', message: string) => void

export function useMachinesWorkspace(props: MachinesProps, emit: MachinesEmit) {
  const { busy, runAction } = useWorkspaceActions()
  const toast = useToast()
  const machines = ref<MachineRecord[]>([])
  const selectedId = ref<string | null>(null)
  const machineFilter = ref<MachineFilter>('All')
  const searchQuery = ref('')
  const editorMode = ref<MachineEditorMode>(null)
  const form = ref<MachineForm>(emptyForm())
  const isLoading = ref(false)
  const isSaving = ref(false)

  const selectedMachine = computed(() => machines.value.find((machine) => machine.id === selectedId.value) ?? null)
  const filteredMachines = computed(() => {
    const query = searchQuery.value.trim().toLowerCase()
    return machines.value.filter((machine) => {
      const matchesFilter = machineFilter.value === 'All' || (machineFilter.value === 'Active' ? machine.active : !machine.active)
      const matchesQuery = !query || [machine.name, machine.code, machine.category, machine.rateBasis, ...(machine.rates || []).map((rate) => `${rate.name} ${rate.selectorValue}`)].some((value) => value.toLowerCase().includes(query))
      return matchesFilter && matchesQuery
    })
  })

  onMounted(loadMachines)

  watch(
    () => props.currencyUnit,
    () => {
      if (editorMode.value !== 'edit' || !selectedMachine.value) return
      form.value.rate = formatMoneyInput(selectedMachine.value.rateRial, props.currencyUnit)
      form.value.setupCost = formatMoneyInput(selectedMachine.value.setupCostRial, props.currencyUnit)
      form.value.rates = rateForms(selectedMachine.value)
    },
  )

  function emptyForm(): MachineForm {
  return { name: '', code: '', category: '', imagePath: '', rateBasis: 'hour', rate: '', setupCost: '', notes: '', rates: [] }
  }

  function emptyRate(index = 0): MachineRateForm {
    return { id: `rate-${Date.now()}-${index}`, name: `Rate ${index + 2}`, selectorValue: '', rateBasis: 'hour', rate: '', setupCost: '', active: true }
  }

  function rateForms(machine: MachineRecord): MachineRateForm[] {
    const rates = Array.isArray(machine.rates) && machine.rates.length
      ? machine.rates
      : [{ id: 'default', name: 'Standard', selectorValue: '', rateBasis: machine.rateBasis, rateRial: machine.rateRial, setupCostRial: machine.setupCostRial, active: true }]
    return rates.slice(1).map((rate) => ({
      id: rate.id,
      name: rate.name,
      selectorValue: rate.selectorValue,
      rateBasis: rate.rateBasis,
      rate: formatMoneyInput(rate.rateRial, props.currencyUnit),
      setupCost: formatMoneyInput(rate.setupCostRial, props.currencyUnit),
      active: rate.active,
    }))
  }

  async function loadMachines() {
    isLoading.value = true
    try {
      machines.value = await machinesApi.list(true)
      if (!selectedId.value) {
        const firstMachine = machines.value.find((machine) => machine.active) ?? machines.value[0]
        selectedId.value = firstMachine?.id ?? null
      }
    } catch (error) {
      toast.error(message(error, 'Machines could not be loaded.'), 'Machines')
    } finally {
      isLoading.value = false
    }
  }

  function selectMachine(id: string) {
    selectedId.value = id
    editorMode.value = null
  }

  function startCreate() {
    editorMode.value = 'create'
    selectedId.value = null
    form.value = emptyForm()
  }

  function startEdit() {
    const machine = selectedMachine.value
    if (!machine) return
    form.value = {
      name: machine.name,
      code: machine.code,
      category: machine.category,
      imagePath: machine.imagePath || '',
      rateBasis: machine.rateBasis,
      rate: formatMoneyInput(machine.rateRial, props.currencyUnit),
      setupCost: formatMoneyInput(machine.setupCostRial, props.currencyUnit),
      notes: machine.notes,
      rates: rateForms(machine),
    }
    editorMode.value = 'edit'
  }

  function cancelEditor() {
    editorMode.value = null
    if (!selectedId.value) {
      const firstMachine = machines.value.find((machine) => machine.active) ?? machines.value[0]
      selectedId.value = firstMachine?.id ?? null
    }
  }

  function backToMachines() {
    selectedId.value = null
    editorMode.value = null
  }

  function rate(value: string) {
    return parseMoneyInput(value, props.currencyUnit)
  }

  function basisLabel(value: string) {
    return ({ unit: 'Per unit / page', minute: 'Per minute', hour: 'Per hour' } as Record<string, string>)[value] ?? value
  }

  function dateLabel(value: string) {
    try {
      return formatDateTime(value)
    } catch {
      return 'Unknown date'
    }
  }

  async function saveMachine() {
    return runAction(async () => {
      const parsedRate = rate(form.value.rate)
      const parsedSetup = form.value.setupCost.trim() === '' ? 0 : rate(form.value.setupCost)
      if (!form.value.name.trim()) {
        toast.error('Enter a machine name.', 'Machines')
        return
      }
      if (parsedRate === null || parsedSetup === null) {
        toast.error(`Enter whole ${props.currencyUnit.toLowerCase()} amounts.`, 'Machines')
        return
      }
      const parsedRates: MachineRatePayload[] = [{ id: 'default', name: 'Standard', selectorValue: '', rateBasis: form.value.rateBasis, rateRial: parsedRate, setupCostRial: parsedSetup, active: true }]
      for (const [index, item] of form.value.rates.entries()) {
        const itemRate = rate(item.rate)
        const itemSetup = item.setupCost.trim() === '' ? 0 : rate(item.setupCost)
        if (!item.name.trim() || itemRate === null || itemSetup === null) {
          toast.error(`Complete rate profile ${index + 2}.`, 'Machines')
          return
        }
        parsedRates.push({ id: item.id, name: item.name.trim(), selectorValue: item.selectorValue.trim(), rateBasis: item.rateBasis, rateRial: itemRate, setupCostRial: itemSetup, active: item.active })
      }
      isSaving.value = true
      const wasEditing = editorMode.value === 'edit'
      try {
        const payload: MachinePayload = { name: form.value.name.trim(), code: form.value.code.trim(), category: form.value.category.trim(), imagePath: form.value.imagePath.trim(), rateBasis: form.value.rateBasis, rateRial: parsedRate, setupCostRial: parsedSetup, notes: form.value.notes.trim(), rates: parsedRates }
        const saved = wasEditing && selectedId.value ? await machinesApi.update(selectedId.value, payload) : await machinesApi.create(payload)
        const index = machines.value.findIndex((item) => item.id === saved.id)
        if (index >= 0) machines.value.splice(index, 1, saved)
        else machines.value.push(saved)
        selectedId.value = saved.id
        editorMode.value = null
        emit('notify', wasEditing ? 'Machine updated.' : 'Machine created.')
      } catch (error) {
        toast.error(message(error, 'Machine could not be saved.'), 'Machines')
      } finally {
        isSaving.value = false
      }
    })
  }

  async function setActive(active: boolean) {
    return runAction(async () => {
      const machine = selectedMachine.value
      if (!machine) return
      try {
        const updated = active ? await machinesApi.reactivate(machine.id) : await machinesApi.archive(machine.id)
        const index = machines.value.findIndex((item) => item.id === updated.id)
        if (index >= 0) machines.value.splice(index, 1, updated)
        emit('notify', active ? 'Machine reactivated.' : 'Machine archived.')
      } catch (error) {
        toast.error(message(error, 'Machine status could not be changed.'), 'Machines')
      }
    })
  }

  async function remove() {
    return runAction(async () => {
      const machine = selectedMachine.value
      if (!machine || !(await confirmAction({ title: 'Remove machine', message: 'Remove this machine permanently when it has no production or service history?', confirmLabel: 'Remove machine', danger: true }))) return
      try {
        await machinesApi.remove(machine.id)
        machines.value = machines.value.filter((item) => item.id !== machine.id)
        selectedId.value = machines.value.find((item) => item.active)?.id ?? machines.value[0]?.id ?? null
        editorMode.value = null
        emit('notify', 'Machine removed.')
      } catch (error) {
        if (!isDeletionProtected(error)) {
          toast.error(message(error, 'Machine could not be removed.'), 'Machines')
          return
        }
        try {
          const archived = await machinesApi.archive(machine.id)
          const index = machines.value.findIndex((item) => item.id === archived.id)
          if (index >= 0) machines.value.splice(index, 1, archived)
          emit('notify', 'Machine is in use, so it was archived instead.')
        } catch (archiveError) {
          toast.error(message(archiveError, 'Machine could not be removed or archived.'), 'Machines')
        }
      }
    })
  }

  function isDeletionProtected(error: unknown) {
    const detail = message(error, '').toLowerCase()
    return detail.includes('archive it instead') || detail.includes('delete protected')
  }

  function message(error: unknown, fallback: string) {
    return error instanceof Error && error.message ? error.message : typeof error === 'string' ? error : fallback
  }

  return { busy, machines, selectedId, selectedMachine, machineFilter, searchQuery, editorMode, form, isLoading, isSaving, filteredMachines, emptyRate, selectMachine, startCreate, startEdit, cancelEditor, backToMachines, saveMachine, setActive, remove, basisLabel, dateLabel }
}
