export function ListMaterials(includeArchived) {
  return window['go']['main']['App']['ListMaterials'](includeArchived)
}

export function GetMaterial(id) {
  return window['go']['main']['App']['GetMaterial'](id)
}

export function CreateMaterial(input) {
  return window['go']['main']['App']['CreateMaterial'](input)
}

export function UpdateMaterial(id, input) {
  return window['go']['main']['App']['UpdateMaterial'](id, input)
}

export function ArchiveMaterial(id) {
  return window['go']['main']['App']['ArchiveMaterial'](id)
}

export function ReactivateMaterial(id) {
  return window['go']['main']['App']['ReactivateMaterial'](id)
}
