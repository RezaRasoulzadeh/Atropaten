import { ArchiveCustomer, CreateCustomer, DeleteCustomer, GetCustomer, ListCustomers, ReactivateCustomer, UpdateCustomer } from '../../wailsjs/go/main/App'
export interface CustomerRecord { id:string; name:string; phone:string; email:string; address:string; notes:string; active:boolean; createdAt:string; updatedAt:string }
export interface CustomerPayload { name:string; phone:string; email:string; address:string; notes:string }
export const customersApi = {
  list(includeArchived = true): Promise<CustomerRecord[]> { return ListCustomers(includeArchived) as Promise<CustomerRecord[]> },
  get(id:string): Promise<CustomerRecord> { return GetCustomer(id) as Promise<CustomerRecord> },
  create(input:CustomerPayload): Promise<CustomerRecord> { return CreateCustomer(input) as Promise<CustomerRecord> },
  update(id:string,input:CustomerPayload): Promise<CustomerRecord> { return UpdateCustomer(id,input) as Promise<CustomerRecord> },
  archive(id:string): Promise<CustomerRecord> { return ArchiveCustomer(id) as Promise<CustomerRecord> },
  reactivate(id:string): Promise<CustomerRecord> { return ReactivateCustomer(id) as Promise<CustomerRecord> },
  remove(id:string): Promise<void> { return DeleteCustomer(id) },
}
