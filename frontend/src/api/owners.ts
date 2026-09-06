import { ArchiveOwner, CloseFiscalPeriod, CreateFiscalPeriod, CreateOwner, CreateOwnerTransaction, DeleteOwner, ListFiscalPeriods, ListOwnerTransactions, ListOwners, PreviewFiscalPeriod, ReactivateOwner, ReverseOwnerTransaction, UpdateOwnerShares } from '../../wailsjs/go/main/App'

export interface OwnerRecord { id:string; name:string; phone:string; email:string; notes:string; active:boolean; ownershipBps:number; profitSharingBps:number; capitalContributedRial:number; drawingsRial:number; currentBalanceRial:number; loanPayableRial:number; loanReceivableRial:number; allocatedProfitLossRial:number; createdAt:string; updatedAt:string }
export interface OwnerTransactionRecord { id:string; transactionNumber:string; ownerId:string; type:string; financialAccountId:string; categoryAccountId:string; description:string; notes:string; status:string; journalEntryId:string; idempotencyKey:string; occurredAt:string; createdAt:string; updatedAt:string; amountRial:number }
export interface AllocationRecord { id:string; periodId:string; ownerId:string; position:number; profitSharingBps:number; amountRial:number }
export interface FiscalPeriodRecord { id:string; name:string; status:string; notes:string; closingJournalEntryId:string; idempotencyKey:string; startDate:string; endDate:string; closedAt:string; createdAt:string; updatedAt:string; revenueRial:number; cogsRial:number; expensesRial:number; profitLossRial:number; allocations:AllocationRecord[]; previewAllocations:AllocationRecord[] }
export const ownersApi={
  list(activeOnly=true){return ListOwners(activeOnly) as unknown as Promise<OwnerRecord[]>},
  create(v:any){return CreateOwner(v) as unknown as Promise<OwnerRecord>},
  updateShares(id:string,v:any){return UpdateOwnerShares(id,v) as unknown as Promise<OwnerRecord>},
  archive(id:string){return ArchiveOwner(id)}, reactivate(id:string){return ReactivateOwner(id)}, delete(id:string){return DeleteOwner(id)},
  transactions(ownerId=''){return ListOwnerTransactions(ownerId) as unknown as Promise<OwnerTransactionRecord[]>},
  createTransaction(v:any){return CreateOwnerTransaction(v) as unknown as Promise<OwnerTransactionRecord>},
  reverseTransaction(id:string,key=''){return ReverseOwnerTransaction(id,key) as unknown as Promise<OwnerTransactionRecord>},
  periods(){return ListFiscalPeriods() as unknown as Promise<FiscalPeriodRecord[]>},
  createPeriod(v:any){return CreateFiscalPeriod(v) as unknown as Promise<FiscalPeriodRecord>},
  previewPeriod(id:string){return PreviewFiscalPeriod(id) as unknown as Promise<FiscalPeriodRecord>},
  closePeriod(id:string,key=''){return CloseFiscalPeriod(id,key) as unknown as Promise<FiscalPeriodRecord>},
}
