import { ArchiveMachine, CreateMachine, ListMachines, ReactivateMachine, UpdateMachine } from '../../wailsjs/go/main/App'
import type { main as mainTypes } from '../../wailsjs/go/models'

export type MachineRecord = mainTypes.MachineDTO
export type MachinePayload = {
  name: string
  code: string
  category: string
  rateBasis: string
  rateRial: number
  setupCostRial: number
  notes: string
}

export const machinesApi = {
  list(includeArchived = true): Promise<MachineRecord[]> { return ListMachines(includeArchived) },
  create(input: MachinePayload): Promise<MachineRecord> { return CreateMachine(input as unknown as mainTypes.MachineInput) },
  update(id: string, input: MachinePayload): Promise<MachineRecord> { return UpdateMachine(id, input as unknown as mainTypes.MachineInput) },
  archive(id: string): Promise<MachineRecord> { return ArchiveMachine(id) },
  reactivate(id: string): Promise<MachineRecord> { return ReactivateMachine(id) },
}
