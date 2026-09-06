import { CreateCheck, DeleteDraftCheck, GetCheck, ListCheckEvents, ListChecks, TransitionCheck } from '../../wailsjs/go/main/App'
export interface CheckRecord { id:string; checkNumber:string; direction:string; bank:string; branch:string; accountDescriptor:string; payerPayee:string; customerId:string; supplierId:string; sourceType:string; sourceId:string; financialAccountId:string; notes:string; status:string; issueDate:string; dueDate:string; createdAt:string; updatedAt:string; amountRial:number }
export interface CheckEventRecord { id:string; checkId:string; fromStatus:string; toStatus:string; note:string; journalEntryId:string; occurredAt:string }
export interface CheckPayload { id?:string; checkNumber:string; direction:string; bank:string; branch?:string; accountDescriptor?:string; payerPayee:string; customerId?:string; supplierId?:string; sourceType?:string; sourceId?:string; financialAccountId?:string; notes?:string; issueDate?:string; dueDate?:string; status?:string; amountRial:number }
export const checksApi={
  list(direction='',status=''){return ListChecks(direction,status) as unknown as Promise<CheckRecord[]>},
  get(id:string){return GetCheck(id) as unknown as Promise<CheckRecord>},
  create(v:CheckPayload){return CreateCheck(v as any) as unknown as Promise<CheckRecord>},
  transition(id:string,to:string,note='',key=''){return TransitionCheck(id,to,note,key) as unknown as Promise<CheckRecord>},
  history(id:string){return ListCheckEvents(id) as unknown as Promise<CheckEventRecord[]>},
  deleteDraft(id:string){return DeleteDraftCheck(id)},
}
