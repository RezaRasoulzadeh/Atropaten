import { CreateLoan, CreateLoanPayment, GetLoan, ListLoanPayments, ListLoans, ReverseLoanPayment } from '../../wailsjs/go/main/App'
export interface LoanInstallmentRecord { id:string; position:number; dueDate:string; principalRial:number; interestFeeRial:number; totalDueRial:number; paidPrincipalRial:number; paidInterestRial:number; paidRial:number; remainingRial:number; overdueRial:number; status:string }
export interface LoanRecord { id:string; loanNumber:string; direction:string; counterpartyName:string; customerId:string; supplierId:string; startDate:string; endDate:string; status:string; notes:string; financialAccountId:string; journalEntryId:string; idempotencyKey:string; createdAt:string; updatedAt:string; principalRial:number; interestFeeRial:number; paidPrincipalRial:number; paidInterestRial:number; remainingPrincipalRial:number; remainingInterestRial:number; overdueRial:number; installments:LoanInstallmentRecord[] }
export interface LoanPayload { id?:string; direction:string; counterpartyName:string; customerId?:string; supplierId?:string; startDate?:string; endDate?:string; notes?:string; financialAccountId:string; idempotencyKey?:string; principalRial:number; interestFeeRial:number; installmentCount?:number; installments?:{id?:string;dueDate:string;principalRial:number;interestFeeRial:number}[] }
export interface LoanPaymentRecord { id:string; paymentNumber:string; loanId:string; financialAccountId:string; paidAt:string; notes:string; status:string; journalEntryId:string; idempotencyKey:string; amountRial:number; principalRial:number; interestRial:number; allocations:{id:string;paymentId:string;installmentId:string;position:number;principalRial:number;interestRial:number}[] }
export interface LoanPaymentPayload { id?:string; loanId:string; financialAccountId:string; paidAt?:string; notes?:string; idempotencyKey?:string; amountRial:number; principalRial:number; interestRial:number; allocations:{installmentId:string;principalRial:number;interestRial:number}[] }
export const loansApi={
  list(direction='',status=''){return ListLoans(direction,status) as unknown as Promise<LoanRecord[]>},
  get(id:string){return GetLoan(id) as unknown as Promise<LoanRecord>},
  create(v:LoanPayload){return CreateLoan(v as any) as unknown as Promise<LoanRecord>},
  payments(id:string){return ListLoanPayments(id) as unknown as Promise<LoanPaymentRecord[]>},
  createPayment(v:LoanPaymentPayload){return CreateLoanPayment(v as any) as unknown as Promise<LoanPaymentRecord>},
  reversePayment(id:string,key=''){return ReverseLoanPayment(id,key) as unknown as Promise<LoanPaymentRecord>},
}
