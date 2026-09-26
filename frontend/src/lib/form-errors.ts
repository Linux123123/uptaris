export function fieldErrorMessages(errors: unknown[]) {
  return errors.map((error) => ({
    message:
      error && typeof error === "object" && "message" in error
        ? String(error.message)
        : String(error),
  }));
}
