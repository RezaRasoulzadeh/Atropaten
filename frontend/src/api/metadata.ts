import {
  AddAttachment,
  CreateProof,
  ImportAttachment,
  ListAttachments,
  ListProofs,
  ReadAttachment,
  RemoveAttachment,
  SaveAttachmentAs,
  UpdateProofStatus,
} from '../../wailsjs/go/main/App'
export interface AttachmentRecord {
  id: string
  ownerType: string
  ownerId: string
  fileName: string
  path: string
  mimeType: string
  sizeBytes: number | null
  checksum: string
  category: string
  notes: string
  createdAt: string
}
export interface AttachmentPreviewRecord {
  fileName: string
  mimeType: string
  contentBase64: string
}
export interface ProofRecord {
  id: string
  ownerType: string
  ownerId: string
  attachmentId: string
  status: string
  versionLabel: string
  approverNote: string
  internalNote: string
  preparedAt: string | null
  approvedAt: string | null
  rejectedAt: string | null
  createdAt: string | null
}
export const metadataApi = {
  attachments(ownerType: string, ownerId: string): Promise<AttachmentRecord[]> {
    return ListAttachments(ownerType, ownerId) as unknown as Promise<AttachmentRecord[]>
  },
  readAttachment(id: string): Promise<AttachmentPreviewRecord> {
    return ReadAttachment(id) as unknown as Promise<AttachmentPreviewRecord>
  },
  saveAttachment(id: string): Promise<boolean> {
    return SaveAttachmentAs(id) as unknown as Promise<boolean>
  },
  addAttachment(input: {
    ownerType: string
    ownerId: string
    fileName: string
    path: string
    mimeType: string
    sizeBytes: number | null
    checksum: string
    category: string
    notes: string
  }): Promise<AttachmentRecord> {
    return AddAttachment(
      input.ownerType,
      input.ownerId,
      input.fileName,
      input.path,
      input.mimeType,
      input.sizeBytes,
      input.checksum,
      input.category,
      input.notes,
    ) as unknown as Promise<AttachmentRecord>
  },
  importAttachment(input: {
    ownerType: string
    ownerId: string
    fileName: string
    mimeType: string
    contentBase64: string
    category: string
    notes: string
  }): Promise<AttachmentRecord> {
    return ImportAttachment(
      input.ownerType,
      input.ownerId,
      input.fileName,
      input.mimeType,
      input.contentBase64,
      input.category,
      input.notes,
    ) as unknown as Promise<AttachmentRecord>
  },
  removeAttachment(id: string): Promise<void> {
    return RemoveAttachment(id) as unknown as Promise<void>
  },
  proofs(ownerType: string, ownerId: string): Promise<ProofRecord[]> {
    return ListProofs(ownerType, ownerId) as unknown as Promise<ProofRecord[]>
  },
  createProof(input: {
    ownerType: string
    ownerId: string
    attachmentId: string
    status: string
    versionLabel: string
    approverNote: string
    internalNote: string
  }): Promise<ProofRecord> {
    return CreateProof(
      input.ownerType,
      input.ownerId,
      input.attachmentId,
      input.status,
      input.versionLabel,
      input.approverNote,
      input.internalNote,
    ) as unknown as Promise<ProofRecord>
  },
  transitionProof(
    ownerType: string,
    ownerId: string,
    proofId: string,
    status: string,
    approverNote: string,
  ): Promise<ProofRecord> {
    return UpdateProofStatus(
      ownerType,
      ownerId,
      proofId,
      status,
      approverNote,
    ) as unknown as Promise<ProofRecord>
  },
}
