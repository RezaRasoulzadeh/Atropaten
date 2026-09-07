import { CreateBackup, GetDataPaths, GetDashboard, GetLastBackup, GetPrintDocument, GetReport, GetShopSettings, RestoreBackup, SaveShopSettings, SelectBackupFile, VerifyBackup } from '../../wailsjs/go/main/App'

export interface ReportSummary { key:string; label:string; amountRial:number; secondaryAmountRial:number; count:number }
export interface ReportRow { id:string; referenceId:string; name:string; secondaryName:string; category:string; status:string; date:string; amountRial:number; secondaryAmountRial:number; tertiaryAmountRial:number; quantityUnits:number; secondaryQuantityUnits:number }
export interface ReportRecord { kind:string; startDate:string; endDate:string; summaries:ReportSummary[]; rows:ReportRow[] }
export interface DashboardRecord { startDate:string; endDate:string; revenueRial:number; grossProfitRial:number; cashRial:number; bankRial:number; receivableRial:number; payableRial:number; openInvoiceCount:number; dueOrderCount:number; overdueOrderCount:number; inProductionCount:number; readyOrderCount:number; attention:any[]; lowStock:any[]; production:any[]; recentActivity:any[] }
export interface ShopSettingsRecord { shopName:string; shopSubtitle:string; phone:string; address:string; email:string; website:string; registrationId:string; taxId:string; logoPath:string; documentFooter:string; documentNotes:string }
export interface PrintDocumentRecord { kind:string; number:string; date:string; dueDate:string; status:string; customerName:string; customerContact:string; supplierName:string; reference:string; method:string; accountName:string; paymentStatus:string; notes:string; subtotalRial:number; discountRial:number; totalRial:number; paidRial:number; remainingRial:number; amountRial:number; shop:ShopSettingsRecord; lines:any[]; statementLines:any[]; allocations:any[] }
export interface DataPathsRecord { root:string; database:string; attachments:string; backups:string; applicationVersion:string; schemaVersion:number }
export interface BackupInfoRecord { path:string; createdAt:string; applicationVersion:string; schemaVersion:number; sizeBytes:number; managedFileCount:number }
export const reportsApi={
  report(kind:string,start:string,end:string){return GetReport(kind,start,end) as unknown as Promise<ReportRecord>},
  dashboard(start:string,end:string){return GetDashboard(start,end) as unknown as Promise<DashboardRecord>},
  print(kind:string,id:string,start:string,end:string,partyId:string){return GetPrintDocument(kind,id,start,end,partyId) as unknown as Promise<PrintDocumentRecord>},
  settings(){return GetShopSettings() as unknown as Promise<ShopSettingsRecord>},
  saveSettings(v:ShopSettingsRecord){return SaveShopSettings(v as any)},
  dataPaths(){return GetDataPaths() as unknown as Promise<DataPathsRecord>},
  lastBackup(){return GetLastBackup() as unknown as Promise<BackupInfoRecord>},
  createBackup(){return CreateBackup() as unknown as Promise<BackupInfoRecord>},
  verifyBackup(path:string){return VerifyBackup(path) as unknown as Promise<BackupInfoRecord>},
  restoreBackup(path:string){return RestoreBackup(path) as unknown as Promise<BackupInfoRecord>},
  selectBackupFile(){return SelectBackupFile()},
}
