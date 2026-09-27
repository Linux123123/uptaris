export function validSecondFactor(code: string, backup: boolean) {
  return backup ? /^[a-f0-9]{20}$/i.test(code.trim().replaceAll("-", "")) : /^\d{6}$/.test(code);
}
