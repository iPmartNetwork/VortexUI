export function sessionRoleLabel(sudo: boolean): string {
  return sudo ? "Super Admin" : "Reseller";
}
