import { CreateExpense, CreatePayment, CreateTransfer, ListAccounts, ListExpenses, ListFinancialAccounts, ListJournalEntries, ListPayments, ListTransfers, ReverseExpense, ReversePayment, ReverseTransfer } from '../../wailsjs/go/main/App'
export interface AccountRecord { id:string; code:string; name:string; type:string; active:boolean; system:boolean; balanceRial:number }
export interface FinancialAccountRecord { id:string; name:string; type:string; ledgerAccountId:string; details:string; active:boolean; balanceRial:number }
export interface JournalLineRecord { id:string; accountId:string; partyType:string; partyId:string; memo:string; position:number; debitRial:number; creditRial:number }
export interface JournalEntryRecord { id:string; entryNumber:string; description:string; sourceType:string; sourceId:string; reversalOfId:string; postedAt:string; createdAt:string; lines:JournalLineRecord[] }
export interface PaymentAllocationRecord { id:string; targetType:string; targetId:string; position:number; amountRial:number; reversed:boolean }
export interface PaymentRecord { id:string; paymentNumber:string; direction:string; method:string; financialAccountId:string; customerId:string; supplierId:string; reference:string; notes:string; status:string; journalEntryId:string; postedAt:string; createdAt:string; amountRial:number; allocations:PaymentAllocationRecord[] }
export interface PaymentPayload { id?:string; direction:string; method:string; financialAccountId:string; customerId?:string; supplierId?:string; reference?:string; notes?:string; postedAt?:string; idempotencyKey?:string; amountRial:number; allocations:{targetType:string;targetId:string;amountRial:number}[] }
export interface ExpenseRecord { id:string; expenseNumber:string; expenseDate:string; categoryAccountId:string; payee:string; supplierId:string; description:string; amountRial:number; paymentMethod:string; financialAccountId:string; notes:string; status:string; journalEntryId:string }
export interface ExpensePayload { id?:string; expenseDate?:string; categoryAccountId:string; payee?:string; supplierId?:string; description:string; amountRial:number; paymentMethod:string; financialAccountId:string; notes?:string; idempotencyKey?:string }
export interface TransferRecord { id:string; transferNumber:string; sourceFinancialAccountId:string; destinationFinancialAccountId:string; amountRial:number; transferDate:string; reference:string; notes:string; status:string; journalEntryId:string }
export interface TransferPayload { id?:string; sourceFinancialAccountId:string; destinationFinancialAccountId:string; amountRial:number; transferDate?:string; reference?:string; notes?:string; idempotencyKey?:string }
export const accountingApi={
  accounts(){return ListAccounts() as unknown as Promise<AccountRecord[]>},
  financialAccounts(){return ListFinancialAccounts() as unknown as Promise<FinancialAccountRecord[]>},
  journal(){return ListJournalEntries() as unknown as Promise<JournalEntryRecord[]>},
  payments(){return ListPayments() as unknown as Promise<PaymentRecord[]>},
  createPayment(v:PaymentPayload){return CreatePayment(v as any) as unknown as Promise<PaymentRecord>},
  reversePayment(id:string,key=''){return ReversePayment(id,key) as unknown as Promise<PaymentRecord>},
  expenses(){return ListExpenses() as unknown as Promise<ExpenseRecord[]>},
  createExpense(v:ExpensePayload){return CreateExpense(v as any) as unknown as Promise<ExpenseRecord>},
  reverseExpense(id:string,key=''){return ReverseExpense(id,key) as unknown as Promise<ExpenseRecord>},
  transfers(){return ListTransfers() as unknown as Promise<TransferRecord[]>},
  createTransfer(v:TransferPayload){return CreateTransfer(v as any) as unknown as Promise<TransferRecord>},
  reverseTransfer(id:string,key=''){return ReverseTransfer(id,key) as unknown as Promise<TransferRecord>},
}
