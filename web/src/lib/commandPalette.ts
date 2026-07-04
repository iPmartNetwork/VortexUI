export function openCommandPalette() {
  window.dispatchEvent(
    new KeyboardEvent("keydown", { key: "k", ctrlKey: true, metaKey: true, bubbles: true }),
  );
}
