import { CreatePayment, ListAccounts, ListFinancialAccounts, ListJournalEntries, ListPayments, ReversePayment } from '../../wailsjs/go/main/App'
export interface AccountRecord { id:string; code:string; name:string; type:string; active:boolean; system:boolean; balanceRial:number }
export interface FinancialAccountRecord { id:string; name:string; type:string; ledgerAccountId:string; details:string; active:boolean; balanceRial:number }
export interface JournalLineRecord { id:string; accountId:string; partyType:string; partyId:string; memo:string; position:number; debitRial:number; creditRial:number }
export interface JournalEntryRecord { id:string; entryNumber:string; description:string; sourceType:string; sourceId:string; reversalOfId:string; postedAt:string; createdAt:string; lines:JournalLineRecord[] }
export interface PaymentAllocationRecord { id:string; targetType:string; targetId:string; position:number; amountRial:number; reversed:boolean }
export interface PaymentRecord { id:string; paymentNumber:string; direction:string; method:string; financialAccountId:string; customerId:string; supplierId:string; reference:string; notes:string; status:string; journalEntryId:string; postedAt:string; createdAt:string; amountRial:number; allocations:PaymentAllocationRecord[] }
export interface PaymentPayload { id?:string; direction:string; method:string; financialAccountId:string; customerId?:string; supplierId?:string; reference?:string; notes?:string; postedAt?:string; idempotencyKey?:string; amountRial:number; allocations:{targetType:string;targetId:string;amountRial:number}[] }
export const accountingApi={
  accounts(){return ListAccounts() as unknown as Promise<AccountRecord[]>},
  financialAccounts(){return ListFinancialAccounts() as unknown as Promise<FinancialAccountRecord[]>},
  journal(){return ListJournalEntries() as unknown as Promise<JournalEntryRecord[]>},
  payments(){return ListPayments() as unknown as Promise<PaymentRecord[]>},
  createPayment(v:PaymentPayload){return CreatePayment(v as any) as unknown as Promise<PaymentRecord>},
  reversePayment(id:string,key=''){return ReversePayment(id,key) as unknown as Promise<PaymentRecord>},
}
